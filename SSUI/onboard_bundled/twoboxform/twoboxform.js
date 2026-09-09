document.addEventListener('DOMContentLoaded', () => {
    const formBackground = document.querySelector('.form-background');
    if (formBackground) {
        const image = new Image();
        const revealBackground = () => document.body.classList.add('form-background-ready');
        image.addEventListener('load', revealBackground, { once: true });
        image.addEventListener('error', revealBackground, { once: true });
        image.src = '/static/login-background.webp';
        if (image.complete) revealBackground();
    }

    // Notification function
    function showNotification(message, type = 'error') {
        const existingNotification = document.querySelector('.notification');
        if (existingNotification) existingNotification.remove();
        const notification = document.createElement('div');
        notification.classList.add('notification', type);
        notification.textContent = message;
        document.body.appendChild(notification);
        notification.offsetHeight;
        notification.classList.add('show');
        setTimeout(() => {
            notification.classList.remove('show');
            setTimeout(() => notification.remove(), 500);
        }, 3000);
    }

    // Preloader functions
    function showPreloader() {
        document.getElementById('preloader').classList.add('show');
    }
    function hidePreloader() {
        document.getElementById('preloader').classList.remove('show');
    }
    async function preloadNextPage() {
        try {
            const response = await fetch('/static/favicon.ico', { method: 'HEAD', cache: 'force-cache' });
            return response.ok;
        } catch (error) {
            console.error('Preload failed:', error);
            return false;
        }
    }

    // Convert yes/no, true/false, 1/0 to boolean strings for config
    function booleanToConfig(value) {
        if (typeof value === 'string') {
            value = value.trim().toLowerCase();
            if (value === 'yes' || value === 'true' || value === '1' || value === 'ja') {
                return true;
            } else if (value === 'no' || value === 'false' || value === '0' || value === 'nej') {
                return false;
            }
        }
        return false; // Default to false if invalid input
    }

    function settingsEndpoint() {
        return document.cookie.includes('SSUICSRF=') ? '/api/v3/settings' : '/api/v3/setup/settings';
    }

    function settingsMethod() {
        return document.cookie.includes('SSUICSRF=') ? 'PATCH' : 'POST';
    }

    function settingsFieldName(name) {
        const aliases = {
            WorldID: 'worldId',
            UPNPEnabled: 'upnpEnabled',
            IsSSCMEnabled: 'sscmEnabled',
            LanguageSetting: 'language'
        };
        return aliases[name] || name.charAt(0).toLowerCase() + name.slice(1);
    }

    // Form submission
    const form = document.getElementById('two-box-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const step = document.getElementById('step').value;
        const mode = document.getElementById('mode').value;
        const configField = document.getElementById('config-field').value;
        let nextStep = document.getElementById('next-step').value;

        if (step === "welcome" || step === "pls_read") {
            window.location.href = `/setup?step=${nextStep}`;
            return;
        }

        if (step === "finalize") {
            // Return to first setup step
            window.location.href = `/setup?step=${nextStep}`;
            return;
        }

        // Handle branching based on Discord enabled/disabled
        if (step === "discord_enabled") {
            const userInput = document.getElementById('primary-field').value.trim().toLowerCase();
            nextStep = (userInput === 'yes' || userInput === 'true' || userInput === '1') 
                ? 'discord_token' 
                : 'network_config_choice'; // Skip Discord setup if not enabled
        }
        
        // Handle branching based on network config choice
        if (step === "network_config_choice") {
            const userInput = document.getElementById('primary-field').value.trim().toLowerCase();
            nextStep = (userInput === 'yes' || userInput === 'true' || userInput === '1') 
                ? 'game_port' 
                : 'admin_account'; // Skip network config if not desired
            
            // No need to save this choice to config
            if (nextStep === 'admin_account') {
                // Skip directly without saving
                window.location.href = `/setup?step=${nextStep}`;
                return;
            } else if (nextStep === 'game_port') {
                // Also skip without saving but go to game port config
                window.location.href = `/setup?step=${nextStep}`;
                return;
            }
        }

        let url, body;
        
        // Handle setup steps
        if (configField && step !== "admin_account") {
            url = settingsEndpoint();
            
            // Handle boolean conversion for yes/no fields
            if (configField === "IsDiscordEnabled" || configField === "UPNPEnabled" || 
                configField === "ServerVisible" || configField === "UseSteamP2P" || configField === "IsSSCMEnabled" || configField === "IsNewTerrainAndSaveSystem") {
                body = JSON.stringify({
                    [settingsFieldName(configField)]: booleanToConfig(document.getElementById('primary-field').value)
                });
                
            } else if (configField === "WorldID") {
                const secondaryValue = document.getElementById('secondary-field').value.trim();
                if (secondaryValue === '' || secondaryValue === document.getElementById('secondary-field').placeholder) {
                    showNotification('Please select a world type!', 'error');
                    hidePreloader();
                    return; // Prevent submission
                    }
                body = JSON.stringify({
                    [settingsFieldName(configField)]: secondaryValue
                });
            } else if (configField === "gameBranch") {
                const secondaryValue = document.getElementById('secondary-field').value.trim();
                if (secondaryValue === '' || secondaryValue === document.getElementById('secondary-field').placeholder) {
                    showNotification('Please select a branch!', 'error');
                    hidePreloader();
                    return; // Prevent submission
                    }
                body = JSON.stringify({
                    [settingsFieldName(configField)]: secondaryValue
                });
            } else {
                body = JSON.stringify({
                    [settingsFieldName(configField)]: document.getElementById('primary-field').value
                });
            }
        } else if (step === "admin_account") { // User setup
            url = '/api/v3/auth/setup/bootstrap';
            body = JSON.stringify({
				username: document.getElementById('primary-field').value,
                password: document.getElementById('secondary-field').value
            });
        } else { // Login or changeuser
            url = mode === 'changeuser' ? '/api/v3/auth/users' : '/api/v3/auth/login';
            body = JSON.stringify({
                username: document.getElementById('primary-field').value,
                password: document.getElementById('secondary-field').value,
                ...(mode === 'changeuser' ? { groupIds: ['system-owner'] } : {})
            });
        }

        try {
            showPreloader();
            // If we're on a step that doesn't need to save data
            if (!url) {
                hidePreloader();
                window.location.href = `/setup?step=${nextStep}`;
                return;
            }

            const response = await fetch(url, {
                method: url === '/api/v3/settings' ? settingsMethod() : 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: body
            });
            const data = await response.json();
            
            if (response.ok || response.status === 201) {
                if (configField || step === "admin_account") {
                    hidePreloader();
                    showNotification(step === "admin_account" ? 'Admin account saved!' : 'Config saved!', 'success');
                    // Wait for backend response to complete before redirecting
                    try { await response.json(); } catch (e) {} // Ensure backend response is fully processed
                    setTimeout(() => {
                        window.location.href = `/setup?step=${nextStep}`;
                    }, 1000);
                } else if (mode === 'login') {
                    showNotification('Login Successful!', 'success');
                    await preloadNextPage();
                    hidePreloader();
                    window.location.href = '/';
                } else { // changeuser
                    hidePreloader();
                    showNotification(data.message || 'User updated!', 'success');
                    form.reset();
                }
            } else {
                hidePreloader();
                showNotification(data.error || 'Action failed!', 'error');
            }
        } catch (error) {
            hidePreloader();
            console.error('Error:', error);
            showNotification('Something went wrong!', 'error');
        }
    });

    // Skip button
    const skipBtn = document.getElementById('skip-btn');
    if (skipBtn) {
        skipBtn.addEventListener('click', () => {
            const step = document.getElementById('step').value;
            let nextStep = document.getElementById('next-step').value;
            
            if (step === "welcome") {
                window.location.href = '/';
                return;
            }
            
            if (step === "finalize") {
                // Go to login page when skipping from finalize
                showNotification('Setup completed!', 'success');
                setTimeout(() => window.location.href = '/', 1000);
                return;
            }
            
            // Custom skip logic for branching steps
            if (step === "discord_enabled") {
                nextStep = "network_config_choice"; // Skip all Discord setup
            } else if (step === "network_config_choice") {
                nextStep = "admin_account"; // Skip all network config
            }
            
            window.location.href = `/setup?step=${nextStep}`;
        });
    }

    // Finalize button - now appears on the finalize page
    document.addEventListener('click', async (e) => {
        if (e.target && e.target.id === 'finalize-btn') {
            try {
                showPreloader();
                const finalizeResponse = await fetch('/api/v3/setup/finalize', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' }
                });
                const data = await finalizeResponse.json();
                
                if (finalizeResponse.ok) {
                    hidePreloader();
                    showNotification(`${data.message}\n${data.restart_hint}`, 'success');
                    setTimeout(() => window.location.href = '/login', 2000);
                } else {
                    hidePreloader();
                    showNotification(data.error || 'Finalize failed!', 'error');
                }
            } catch (error) {
                hidePreloader();
                console.error('Finalize error:', error);
                showNotification('Error finalizing setup!', 'error');
            }
        }
    });

    // Language flag selection
    const languageFlags = document.querySelectorAll('#language-flags img, #welcome-flags img');
    languageFlags.forEach(flag => {
        flag.addEventListener('click', async () => {
            const lang = flag.dataset.lang;
            try {
                showPreloader();
                const response = await fetch(settingsEndpoint(), {
                    method: settingsMethod(),
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ [settingsFieldName('LanguageSetting')]: lang })
                });
                const data = await response.json();
                if (response.ok) {
                    showNotification(`Language set to ${lang}`, 'success');
                } else {
                    showNotification(data.error || 'Failed to set language', 'error');
                }
            } catch (error) {
                console.error('Language setting error:', error);
                showNotification('Error setting language!', 'error');
            } finally {
                hidePreloader();
            }
            window.location.reload();
        });
    });
});
