package discordbot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/SteamServerUI/StationeersServerUI/v6/src/config"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/backupmgr"
	"github.com/SteamServerUI/StationeersServerUI/v6/src/managers/gamemgr"
	"github.com/bwmarrin/discordgo"
)

type discordTransport func(*http.Request) (*http.Response, error)

func (f discordTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type discordRequest struct {
	method, path, contentType string
	body                      []byte
}

func fakeDiscord(t *testing.T) (*discordgo.Session, *[]discordRequest) {
	t.Helper()
	s, err := discordgo.New("Bot test-token")
	if err != nil {
		t.Fatal(err)
	}
	s.State.User = &discordgo.User{ID: "bot"}
	requests := &[]discordRequest{}
	s.Client = &http.Client{Transport: discordTransport(func(r *http.Request) (*http.Response, error) {
		var body []byte
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
		}
		*requests = append(*requests, discordRequest{r.Method, r.URL.Path, r.Header.Get("Content-Type"), body})
		payload := `{"id":"message","channel_id":"hub"}`
		if r.Method == http.MethodGet {
			payload = `[]`
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(payload)), Request: r}, nil
	})}
	return s, requests
}

func hubTestConfig(t *testing.T) {
	t.Helper()
	config.ConfigMu.Lock()
	enabled, role, channel, session := config.IsDiscordEnabled, config.DiscordAdminRoleID, config.StatusPanelChannelID, config.DiscordSession
	config.IsDiscordEnabled, config.DiscordAdminRoleID, config.StatusPanelChannelID, config.DiscordSession = true, "123", "hub", nil
	config.ConfigMu.Unlock()
	resetHubDialogs()
	t.Cleanup(func() {
		config.ConfigMu.Lock()
		config.IsDiscordEnabled, config.DiscordAdminRoleID, config.StatusPanelChannelID, config.DiscordSession = enabled, role, channel, session
		config.ConfigMu.Unlock()
		resetHubDialogs()
	})
}

func hubInteraction(admin bool) *discordgo.InteractionCreate {
	roles := []string{}
	if admin {
		roles = []string{"123"}
	}
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "interaction", AppID: "bot", Token: "test", GuildID: "guild", ChannelID: "hub", Type: discordgo.InteractionMessageComponent,
		Member: &discordgo.Member{User: &discordgo.User{ID: "alice"}, Roles: roles},
		Data:   discordgo.MessageComponentInteractionData{CustomID: ButtonAdminActions},
	}}
}

func lastResponse(t *testing.T, requests *[]discordRequest) discordgo.InteractionResponse {
	t.Helper()
	if len(*requests) == 0 {
		t.Fatal("no Discord response")
	}
	var wire struct {
		Type discordgo.InteractionResponseType `json:"type"`
		Data json.RawMessage                   `json:"data"`
	}
	if err := json.Unmarshal((*requests)[len(*requests)-1].body, &wire); err != nil {
		t.Fatal(err)
	}
	// Message has DiscordGo's component decoder; response data is send-only.
	var message discordgo.Message
	if err := json.Unmarshal(wire.Data, &message); err != nil {
		t.Fatal(err)
	}
	return discordgo.InteractionResponse{Type: wire.Type, Data: &discordgo.InteractionResponseData{Embeds: message.Embeds, Components: message.Components, Flags: message.Flags}}
}

func TestHubAccessAndPrivateResponses(t *testing.T) {
	hubTestConfig(t)
	for _, admin := range []bool{false, true} {
		s, requests := fakeDiscord(t)
		i := hubInteraction(admin)
		handleAdminInteraction(s, i)
		response := lastResponse(t, requests)
		if response.Data.Flags != discordgo.MessageFlagsEphemeral {
			t.Fatal("public admin response")
		}
		if got := len(response.Data.Components) > 0; got != admin {
			t.Fatalf("admin=%v, menu=%v", admin, got)
		}
	}
	for _, mutate := range []func(*discordgo.InteractionCreate){
		func(i *discordgo.InteractionCreate) {
			i.Member.Roles = nil
			i.Member.Permissions = discordgo.PermissionAdministrator
		},
		func(i *discordgo.InteractionCreate) { i.ChannelID = "elsewhere" },
		func(i *discordgo.InteractionCreate) { i.GuildID = "" },
		func(i *discordgo.InteractionCreate) { i.Member = nil },
	} {
		i := hubInteraction(true)
		mutate(i)
		if hasDiscordAdminRole(i) {
			t.Fatal("accepted unauthorized interaction")
		}
	}
	config.ConfigMu.Lock()
	config.DiscordAdminRoleID = ""
	config.ConfigMu.Unlock()
	if hasDiscordAdminRole(hubInteraction(true)) {
		t.Fatal("missing role fails open")
	}
}

func TestEverySlashCommandExceptStatusRequiresRole(t *testing.T) {
	hubTestConfig(t)
	for _, command := range []string{"start", "stop", "restart", "update", "help", "list", "restore", "download", "announce", "command", "bansteamid", "unbansteamid", "status"} {
		s, requests := fakeDiscord(t)
		i := hubInteraction(false)
		i.Type, i.Data = discordgo.InteractionApplicationCommand, discordgo.ApplicationCommandInteractionData{Name: command}
		listenToSlashCommands(s, i)
		response := lastResponse(t, requests)
		if response.Data.Flags != discordgo.MessageFlagsEphemeral {
			t.Fatalf("%s responded publicly", command)
		}
		denied := response.Data.Embeds[0].Title == "😔 Sorry!"
		if denied != (command != "status") {
			t.Fatalf("wrong access for /%s", command)
		}
	}
}

func TestConfirmationsRecheckRoleAndAreSingleUse(t *testing.T) {
	hubTestConfig(t)
	s, requests := fakeDiscord(t)
	i := hubInteraction(true)
	i.Type, i.Data = discordgo.InteractionApplicationCommand, discordgo.ApplicationCommandInteractionData{Name: "restore", Options: []*discordgo.ApplicationCommandInteractionDataOption{{Name: "name", Type: discordgo.ApplicationCommandOptionString, Value: "original.save"}}}
	listenToSlashCommands(s, i)
	response := lastResponse(t, requests)
	row := response.Data.Components[0].(*discordgo.ActionsRow)
	id := row.Components[0].(*discordgo.Button).CustomID
	if !strings.Contains(response.Data.Embeds[0].Description, "original.save") {
		t.Fatal("confirmation lost its backup")
	}
	i.Type = discordgo.InteractionMessageComponent
	i.Data = discordgo.MessageComponentInteractionData{CustomID: id}
	i.Member.Roles = nil
	handleAdminInteraction(s, i)
	if lastResponse(t, requests).Data.Embeds[0].Title != "😔 Sorry!" {
		t.Fatal("revoked role accepted")
	}
	parts := strings.Split(strings.TrimPrefix(id, hubDialogPrefix), ":")
	i.Member.Roles = []string{"123"}
	var wins atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if _, ok := takeHubDialog(i, parts[0]); ok {
				wins.Add(1)
			}
		})
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("confirmation accepted %d times", wins.Load())
	}
}

func TestHubDialogScopeExpiryAndBound(t *testing.T) {
	hubTestConfig(t)
	i := hubInteraction(true)
	id, _ := storeHubDialog(i, hubDialog{screen: "confirm", action: "stop"})
	other := hubInteraction(true)
	other.Member.User.ID = "bob"
	if _, ok := takeHubDialog(other, id); ok {
		t.Fatal("another user consumed a menu")
	}
	other = hubInteraction(true)
	other.ChannelID = "elsewhere"
	if _, ok := takeHubDialog(other, id); ok {
		t.Fatal("another channel consumed a menu")
	}
	other = hubInteraction(true)
	other.GuildID = "elsewhere"
	if _, ok := takeHubDialog(other, id); ok {
		t.Fatal("another guild consumed a menu")
	}
	if _, ok := takeHubDialog(i, id); !ok {
		t.Fatal("owner lost the menu")
	}
	id, _ = storeHubDialog(i, hubDialog{})
	hubDialogs.Lock()
	d := hubDialogs.items[id]
	d.expires = time.Now().Add(-time.Second)
	hubDialogs.items[id] = d
	hubDialogs.Unlock()
	if _, ok := takeHubDialog(i, id); ok {
		t.Fatal("expired menu accepted")
	}
	for n := range maxHubDialogs {
		i.Member.User.ID = fmt.Sprint(n)
		if _, err := storeHubDialog(i, hubDialog{}); err != nil {
			t.Fatal(err)
		}
	}
	i.Member.User.ID = "extra"
	if _, err := storeHubDialog(i, hubDialog{}); err == nil {
		t.Fatal("menu cache is unbounded")
	}
	resetHubDialogs()
	if _, ok := takeHubDialog(i, id); ok {
		t.Fatal("menu survived runtime reset")
	}
}

func TestBackupPagesPinLongNamesAndStayWithinDiscordLimits(t *testing.T) {
	hubTestConfig(t)
	i := hubInteraction(true)
	backups := make([]backupmgr.BackupSaveFile, 4000)
	for n := range backups {
		backups[n] = backupmgr.BackupSaveFile{Name: fmt.Sprintf("nested/%04d/%s.save", n, strings.Repeat("🌍", 110)), SaveTime: time.Now(), SummaryReady: true}
	}
	for _, page := range []int{0, 1, 399, 9999} {
		embed, components, err := buildBackupPage(i, backups, page, 0)
		if err != nil {
			t.Fatal(err)
		}
		if utf8.RuneCountInString(embed.Description) > 4096 {
			t.Fatal("oversized embed")
		}
		menu := components[0].(discordgo.ActionsRow).Components[0].(discordgo.SelectMenu)
		if len(menu.Options) > 25 || len(menu.CustomID) > 100 {
			t.Fatal("oversized select")
		}
		id := strings.Split(strings.TrimPrefix(menu.CustomID, hubDialogPrefix), ":")[0]
		d, ok := takeHubDialog(i, id)
		if !ok || len(d.names) != 10 || d.names[0] != backups[min(page, 399)*10].Name {
			t.Fatal("page lost its filenames")
		}
		for _, option := range menu.Options {
			if len(option.Value) > 100 || utf8.RuneCountInString(option.Label) > 100 {
				t.Fatal("oversized option")
			}
		}
		// Reordering the current inventory must not alter the displayed selection.
		name := d.names[0]
		backups[0], backups[1] = backups[1], backups[0]
		if d.names[0] != name {
			t.Fatal("selection changed")
		}
	}
	if _, _, err := buildBackupPage(i, nil, 0, 0); err == nil {
		t.Fatal("empty page accepted")
	}
}

func TestHubModalAndInputValidation(t *testing.T) {
	hubTestConfig(t)
	s, requests := fakeDiscord(t)
	i := hubInteraction(true)
	i.Data = discordgo.MessageComponentInteractionData{CustomID: adminSelectID, Values: []string{"command"}}
	handleAdminInteraction(s, i)
	response := lastResponse(t, requests)
	if response.Type != discordgo.InteractionResponseModal {
		t.Fatal("console did not open a modal")
	}
	var data discordgo.ModalSubmitInteractionData
	if err := json.Unmarshal([]byte(`{"custom_id":"modal","components":[{"type":1,"components":[{"type":4,"custom_id":"value","value":"status"}]}]}`), &data); err != nil {
		t.Fatal(err)
	}
	if value, err := hubModalValue(data); err != nil || value != "status" {
		t.Fatalf("modal value: %q %v", value, err)
	}
	for _, test := range []struct{ action, arg string }{{"command", "one\ntwo"}, {"announce", "hello\rquit"}, {"command", strings.Repeat("x", 1001)}, {"bansteamid", "1,2"}, {"bansteamid", strings.Repeat("x", 17)}, {"bad", ""}} {
		if validateAdminInput(test.action, test.arg) == nil {
			t.Fatalf("accepted invalid input: %+v", test)
		}
	}
	if err := validateAdminInput("bansteamid", "76561198000000000"); err != nil {
		t.Fatal(err)
	}
}

func TestServerActionOrderingAndFailures(t *testing.T) {
	for _, test := range []struct {
		action, fail, want string
		running            bool
	}{
		{"start", "", "running,start", false}, {"start", "", "running", true},
		{"stop", "", "running,stop", true}, {"stop", "", "running", false},
		{"restart", "", "running,stop,start", true}, {"restart", "stop", "running,stop", true},
		{"update", "", "running,stop,update", true}, {"update", "update", "running,stop,update", true},
		{"restore", "", "check:chosen.save,running,stop,restore:chosen.save,start", true},
		{"restore", "check", "check:chosen.save", true},
		{"restore", "restore", "check:chosen.save,running,stop,restore:chosen.save", true},
		{"restore", "start", "check:chosen.save,running,restore:chosen.save,start", false},
	} {
		var calls []string
		call := func(label string) error {
			calls = append(calls, label)
			if strings.Split(label, ":")[0] == test.fail {
				return errors.New("test failure")
			}
			return nil
		}
		backend := serverActionBackend{running: func() bool { calls = append(calls, "running"); return test.running }, start: func() error { return call("start") }, stop: func() error { return call("stop") }, update: func() error { return call("update") }, checkBackup: func(name string) error { return call("check:" + name) }, restore: func(name string) error { return call("restore:" + name) }}
		_, err := performServerAction(test.action, "chosen.save", backend)
		if (err != nil) != (test.fail != "") || strings.Join(calls, ",") != test.want {
			t.Fatalf("%+v: calls %v err %v", test, calls, err)
		}
	}
}

func TestDiscordActionsShareOneSlot(t *testing.T) {
	finishDiscordAction(nil)
	t.Cleanup(func() {
		discordAction.Lock()
		discordAction.busy = false
		discordAction.name = ""
		discordAction.Unlock()
	})
	if !beginDiscordAction("restore") || beginDiscordAction("restart vote") {
		t.Fatal("conflicting actions overlapped")
	}
	resetHubDialogs()
	if beginDiscordAction("update") {
		t.Fatal("runtime reset released active operation")
	}
	finishDiscordAction(nil)
	if !beginDiscordAction("restart vote") {
		t.Fatal("finished operation retained slot")
	}
	finishDiscordAction(nil)
}

func TestPrivateDownloadsNeverPostToChannel(t *testing.T) {
	hubTestConfig(t)
	dir := t.TempDir()
	name := "040926_112440_auto.save"
	payload := []byte("private backup contents")
	if err := os.WriteFile(filepath.Join(dir, name), payload, 0600); err != nil {
		t.Fatal(err)
	}
	manager := backupmgr.NewBackupManager(backupmgr.BackupConfig{SafeBackupDir: dir, BackupDir: t.TempDir()})
	s, requests := fakeDiscord(t)
	i := hubInteraction(true)
	if !deferHub(s, i) {
		t.Fatal("defer failed")
	}
	downloadBackupReply(s, i, manager, name)
	if len(*requests) != 2 {
		t.Fatalf("unexpected requests: %d", len(*requests))
	}
	if !bytes.Contains((*requests)[1].body, payload) {
		t.Fatal("missing download attachment")
	}
	for _, request := range *requests {
		if strings.Contains(request.path, "/channels/") {
			t.Fatal("public backup upload")
		}
	}
	if last := (*requests)[1]; last.method != http.MethodPatch || !strings.Contains(last.path, "/webhooks/") {
		t.Fatal("download did not edit private response")
	}
	*requests = nil
	downloadBackupReply(s, i, manager, "missing.save")
	for _, request := range *requests {
		if strings.Contains(request.path, "/channels/") {
			t.Fatal("public failure fallback")
		}
	}
}

func TestStatusPanelUsesRealStateAndBoundedFields(t *testing.T) {
	hubTestConfig(t)
	old := gamemgr.GetServerState()
	t.Cleanup(func() { gamemgr.SetServerState(old) })
	gamemgr.SetServerState(gamemgr.ServerStateStopped)
	players := map[string]string{}
	for n := range 200 {
		players[fmt.Sprint(n)] = strings.Repeat("x", 100)
	}
	embed := buildStatusPanelEmbed(players, &backupmgr.SaveSummary{DaysPlayed: 42, Things: 100, Atmospheres: 200, Rooms: 30, PipeNetworks: 40, CableNetworks: 50, SavedAt: time.Now()})
	if !strings.Contains(embed.Description, "Offline") || embed.Color != 0xED4245 {
		t.Fatal("players turned a stopped server green")
	}
	for _, field := range embed.Fields {
		if utf8.RuneCountInString(field.Value) > 1024 {
			t.Fatal("oversized field")
		}
	}
	for _, name := range []string{"Days played", "Things"} {
		if !slices.ContainsFunc(embed.Fields, func(field *discordgo.MessageEmbedField) bool { return strings.Contains(field.Name, name) }) {
			t.Fatalf("hub is missing %s", name)
		}
	}
	for _, name := range []string{"Atmospheres", "Rooms", "Pipe networks", "Cable networks", "Active Votes"} {
		if slices.ContainsFunc(embed.Fields, func(field *discordgo.MessageEmbedField) bool { return strings.Contains(field.Name, name) }) {
			t.Fatalf("hub still contains %s", name)
		}
	}
	components := buildPanelComponents()
	if len(components) > 5 {
		t.Fatal("too many rows")
	}
	for _, row := range components {
		if len(row.(discordgo.ActionsRow).Components) > 5 {
			t.Fatal("too many buttons")
		}
	}
	if !slices.ContainsFunc(components, func(component discordgo.MessageComponent) bool {
		return slices.ContainsFunc(component.(discordgo.ActionsRow).Components, func(button discordgo.MessageComponent) bool {
			item, ok := button.(discordgo.Button)
			return ok && item.CustomID == ButtonSaveStats
		})
	}) {
		t.Fatal("save stats button is missing")
	}
}

func TestSaveStatsButtonShowsFullCachedSummaryPrivately(t *testing.T) {
	hubTestConfig(t)
	statusPanelData.Lock()
	oldSummary := statusPanelData.summary
	statusPanelData.summary = &backupmgr.SaveSummary{DaysPlayed: 42, Things: 100, Atmospheres: 200, Rooms: 30, PipeNetworks: 40, CableNetworks: 50, SavedAt: time.Now()}
	statusPanelData.Unlock()
	t.Cleanup(func() {
		statusPanelData.Lock()
		statusPanelData.summary = oldSummary
		statusPanelData.Unlock()
	})

	s, requests := fakeDiscord(t)
	i := hubInteraction(false)
	i.Data = discordgo.MessageComponentInteractionData{CustomID: ButtonSaveStats}
	handlePanelButtonInteraction(s, i)
	response := lastResponse(t, requests)
	if response.Data.Flags != discordgo.MessageFlagsEphemeral || response.Data.Embeds[0].Title != "📊 Save Stats" {
		t.Fatal("save stats response was not private")
	}
	for _, name := range []string{"Days played", "Things", "Atmospheres", "Rooms", "Pipe networks", "Cable networks"} {
		if !slices.ContainsFunc(response.Data.Embeds[0].Fields, func(field *discordgo.MessageEmbedField) bool { return strings.Contains(field.Name, name) }) {
			t.Fatalf("save stats are missing %s", name)
		}
	}
}

func TestFinishedDiscordActionDisappearsAfterOneHour(t *testing.T) {
	now := time.Now()
	discordAction.Lock()
	oldBusy, oldName, oldResult, oldChanged := discordAction.busy, discordAction.name, discordAction.result, discordAction.changed
	discordAction.busy, discordAction.name, discordAction.result = false, "update", "Completed"
	discordAction.changed = now.Add(-59 * time.Minute)
	discordAction.Unlock()
	t.Cleanup(func() {
		discordAction.Lock()
		discordAction.busy, discordAction.name, discordAction.result, discordAction.changed = oldBusy, oldName, oldResult, oldChanged
		discordAction.Unlock()
	})

	if discordActionStatusAt(now) == "" {
		t.Fatal("recent action disappeared")
	}
	discordAction.Lock()
	discordAction.changed = now.Add(-61 * time.Minute)
	discordAction.Unlock()
	if discordActionStatusAt(now) != "" {
		t.Fatal("old completed action is still visible")
	}
	discordAction.Lock()
	discordAction.busy = true
	discordAction.Unlock()
	if discordActionStatusAt(now) == "" {
		t.Fatal("long-running action disappeared")
	}
}

func TestHubReusesOnlyOwnMessageAndDoesNotDeleteHistory(t *testing.T) {
	hubTestConfig(t)
	s, _ := fakeDiscord(t)
	var methods []string
	s.Client.Transport = discordTransport(func(r *http.Request) (*http.Response, error) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		payload := `{"id":"ours"}`
		if r.Method == http.MethodGet {
			payload = `[
			{"id":"someone-else","author":{"id":"other-bot"},"embeds":[{"footer":{"text":"SSUI Hub • v5"}}]},
			{"id":"ordinary-chat","author":{"id":"alice"},"content":"Keep me"},
			{"id":"ours","author":{"id":"bot"},"embeds":[{"footer":{"text":"SSUI Hub • v5"}}]}
		]`
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(payload)), Request: r}, nil
	})
	config.ConfigMu.Lock()
	config.DiscordSession = s
	config.ConfigMu.Unlock()
	statusPanelMutex.Lock()
	statusPanelMessageID = ""
	statusPanelChannelID = ""
	statusPanelMutex.Unlock()
	t.Cleanup(func() {
		statusPanelMutex.Lock()
		statusPanelMessageID = ""
		statusPanelChannelID = ""
		statusPanelMutex.Unlock()
	})
	sendOrEditStatusPanel("hub", hubEmbed("Test", "Test", 0), nil)
	sendOrEditStatusPanel("hub", hubEmbed("Test", "Updated", 0), nil)
	if len(methods) != 3 || !strings.HasSuffix(methods[1], "/ours") || !strings.HasPrefix(methods[1], "PATCH ") {
		t.Fatalf("did not adopt hub: %v", methods)
	}
	for _, method := range methods {
		if strings.HasPrefix(method, "DELETE ") || strings.HasPrefix(method, "POST ") {
			t.Fatalf("modified other channel messages: %v", methods)
		}
	}
}

func TestHubDoesNotDuplicatePanelOnDiscordFailure(t *testing.T) {
	hubTestConfig(t)
	for _, status := range []int{403, 404, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			s, _ := fakeDiscord(t)
			s.MaxRestRetries = 0
			posts := 0
			s.Client.Transport = discordTransport(func(r *http.Request) (*http.Response, error) {
				code, payload := status, `{"code":50013,"message":"test failure"}`
				if r.Method == http.MethodPost {
					posts++
					code = 200
					payload = `{"id":"replacement"}`
				}
				return &http.Response{StatusCode: code, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(payload)), Request: r}, nil
			})
			config.ConfigMu.Lock()
			config.DiscordSession = s
			config.ConfigMu.Unlock()
			statusPanelMutex.Lock()
			statusPanelMessageID = "existing"
			statusPanelChannelID = "hub"
			statusPanelMutex.Unlock()
			sendOrEditStatusPanel("hub", hubEmbed("Test", "Test", 0), nil)
			if (posts == 1) != (status == 404) {
				t.Fatalf("status %d created %d panels", status, posts)
			}
		})
	}
	statusPanelMutex.Lock()
	statusPanelMessageID = ""
	statusPanelChannelID = ""
	statusPanelMutex.Unlock()
}

func TestPrivateMenuUpdatesAndOldDownloads(t *testing.T) {
	hubTestConfig(t)
	s, requests := fakeDiscord(t)
	i := hubInteraction(true)
	i.Message = &discordgo.Message{Flags: discordgo.MessageFlagsEphemeral}
	handleAdminInteraction(s, i)
	if lastResponse(t, requests).Type != discordgo.InteractionResponseUpdateMessage {
		t.Fatal("private menu spawned another message")
	}
	i.Data = discordgo.MessageComponentInteractionData{CustomID: ButtonDownloadBackupPfx + "old.save"}
	handleAdminInteraction(s, i)
	response := lastResponse(t, requests)
	if response.Data.Flags != discordgo.MessageFlagsEphemeral || !strings.Contains(response.Data.Embeds[0].Description, "replaced") {
		t.Fatal("old download button remained active")
	}
	for _, request := range *requests {
		if strings.Contains(request.path, "/channels/") {
			t.Fatal("old button sent a public message")
		}
	}
}

func TestBusyActionDoesNotConsumePassedVote(t *testing.T) {
	hubTestConfig(t)
	resetDiscordVotes()
	t.Cleanup(resetDiscordVotes)
	discordVotes.Lock()
	discordVotes.restart = &activeVote{kind: voteRestart, required: 1, voters: map[string]struct{}{}, expiresAt: time.Now().Add(time.Minute)}
	discordVotes.Unlock()
	if !beginDiscordAction("update") {
		t.Fatal("could not reserve action")
	}
	t.Cleanup(func() {
		finishDiscordAction(nil)
		discordAction.Lock()
		discordAction.name = ""
		discordAction.Unlock()
	})
	result := castDiscordVote(voteRestart, restoreVoteTarget{}, "alice")
	if !strings.Contains(result.message, "not added") {
		t.Fatalf("busy vote: %+v", result)
	}
	discordVotes.Lock()
	defer discordVotes.Unlock()
	if discordVotes.restart == nil || len(discordVotes.restart.voters) != 0 || !discordVotes.restartCooldownUntil.IsZero() {
		t.Fatal("busy action consumed the vote or its cooldown")
	}
}
