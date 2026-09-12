!!! danger "⚠️ LEGACY DOCUMENTATION"
    This page describes the old SSUI v5 documentation and is kept for reference only. It does not describe the current SSUI v6 release, API, paths or security model.

    For a new installation, start with the [current v6 documentation](../../index.md).

# Available Commands

| Command | Parameters | Description |
|---------|------------|-------------|
| addgas | [Oxygen,Nitrogen,CarbonDioxide,Volatiles,Pollutant,Water,NitrousOxide] | Adds GasType to target thing |
| atmos | [pipe,world,direction,room,global,thing,cleanup,count,liquid] | Enables atmosphere debugging |
| ban | [<clientId>,refresh] | Bans a client from the server |
| camera | [shake] | Various camera debug functions |
| celestial | [eccentricity,semimajoraxisau,semimajoraxiskm,inclination,periapsis,period,ascendingnode,rotation] | Allows editing of celestial bodies |
| cleanupplayers | [dead,disconnected,all] | Cleans up player bodies |
| clear |  | Clears all console text |
| debugthreads | [GameTick,Terrain] | Show worker thread run times |
| deletelooseitems |  | Removes all loose items in world |
| deleteoutofbounds |  | Removes out-of-bounds objects |
| difficulty | [<?difficulty>] | Prints or sets difficulty |
| discord |  | Interaction with Discord SDK |
| dlc | [shared] | Various DLC debug functions |
| emote | [emoteName] | Triggers player emote |
| entity | [state <playerName OR referenceId>] | Entity debug functions |
| exportworld |  | Exports world to WorldSettings file |
| help | [commands,list,<key>,tofile] | Displays command help |
| helperhints | [Dismiss,Complete,Trigger] | Tests world objectives |
| keybindings | [reset] | Displays or resets keybindings |
| kick | [<clientId>] | Kicks a client from server |
| legacycpu | [enable,disable] | Enables Legacy CPU mode |
| liquid | [show,renderer,solver,WorldVolume] | Debugs liquid solver |
| listnetworkdevices | [id] | Lists network devices |
| localization | [None,WordCount,Generate,Refresh,CheckKeys,CheckFonts] | Displays localization info |
| log | [<logname>,clear] | Dumps logs to file |
| logtoclipboard |  | Copies console to clipboard |
| masterserver | [refresh] | Interacts with Master Server |
| minables | [range,generate] | Toggles minable debug |
| netconfig | [list,print,<PropertyName> <Value>] | Changes NetConfig.xml |
| network |  | Shows network status |
| networkdebug |  | Displays network debug window |
| orbit | [debug,view,celestials,simulate,set,timescale,makeoffset] | Controls orbital simulation |
| pause | [true,false] | Pauses/unpauses game |
| plant | [grow <parent thing id>] | Plant debug functions |
| prefabs | [Thumbnails] | Validates source prefabs |
| printgasinfo |  | Prints gas coefficients |
| profiler | [enable,disable] | Toggles profiler |
| regeneraterooms |  | Regenerates world rooms |
| rocket | [refresh,print,abandon,debug,chart] | Rocket debug functions |
| save | [<filename>,delete <filename>,list] | Saves game |
| say |  | Sends message to players |
| setbatteries | [Empty,Critical,Very Low,Low,Medium,High,Full] | Sets battery levels |
| settings | [list,print,<PropertyName> <Value>] | Changes settings.xml |
| settingspath | [<full-directory-path>] | Sets settings path |
| spacemap | [regenerate,fill,chart,testpaths] | Space map debug functions |
| spacemapnode | [<id>] | Space map node debug |
| status |  | Shows server state |
| steam | [Refresh,Store,Achieve,Clear,ClearAll,Invalid] | Tests Steamworks |
| storm | [start,stop,debug] | Controls weather events |
| structure | [completeall] | Structure debug functions |
| structurenetwork | [chute,rocket] | Debugs structure networks |
| systeminfo |  | Prints system info |
| test |  | Tests colors |
| testbytearray |  | Tests network read/write |
| testoctree | [[number of iterations]] | Benchmarks read density |
| thing | [find <id>,delete <id>,spawn <prefabName> [amount],info <id>,...] | Manages things |
| trader | [regenerate,land,depart,contacts,buys,sells,evaluate,checksum] | Trader debug commands |
| unstuck |  | Attempts to unstick player |
| up
