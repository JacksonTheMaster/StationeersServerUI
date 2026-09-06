document.addEventListener('DOMContentLoaded', () => {
    // Planet creation (unchanged)
    const planetContainer = document.getElementById('planet-container');
    function createPlanet(container, size, orbitRadius, speed, color) {
        const orbit = document.createElement('div');
        orbit.classList.add('orbit');
        orbit.style.width = `${orbitRadius * 2}px`;
        orbit.style.height = `${orbitRadius * 2}px`;
        orbit.style.position = 'absolute';
        orbit.style.left = '50%';
        orbit.style.top = '50%';
        orbit.style.transform = 'translate(-50%, -50%)';
        const randomDelay = -(Math.random() * speed);
        orbit.style.animation = `orbit ${speed}s linear infinite ${randomDelay}s`;
        const planet = document.createElement('div');
        planet.classList.add('planet');
        planet.style.width = `${size}px`;
        planet.style.height = `${size}px`;
        planet.style.position = 'absolute';
        planet.style.left = '0%';
        planet.style.top = '50%';
        planet.style.backgroundColor = color;
        planet.style.borderRadius = '50%';
        planet.style.boxShadow = `0 0 20px ${color}`;
        orbit.appendChild(planet);
        container.appendChild(orbit);
    }
    createPlanet(planetContainer, 80, 650, 30, 'rgba(200, 100, 50, 0.7)');
    createPlanet(planetContainer, 50, 1000, 50, 'rgba(100, 200, 150, 0.5)');
    createPlanet(planetContainer, 30, 1250, 60, 'rgba(50, 150, 250, 0.6)');
    createPlanet(planetContainer, 70, 350, 30, 'rgba(200, 150, 200, 0.7)');

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
            url = '/api/v2/saveconfig';
            
            // Handle boolean conversion for yes/no fields
            if (configField === "IsDiscordEnabled" || configField === "UPNPEnabled" || 
                configField === "ServerVisible" || configField === "UseSteamP2P" || configField === "IsSSCMEnabled" || configField === "IsNewTerrainAndSaveSystem") {
                body = JSON.stringify({
                    [configField]: booleanToConfig(document.getElementById('primary-field').value)
                });
                
            } else if (configField === "WorldID") {
                const secondaryValue = document.getElementById('secondary-field').value.trim();
                if (secondaryValue === '' || secondaryValue === document.getElementById('secondary-field').placeholder) {
                    showNotification('Please select a world type!', 'error');
                    hidePreloader();
                    return; // Prevent submission
                    }
                body = JSON.stringify({
                    [configField]: secondaryValue
                });
            } else if (configField === "gameBranch") {
                const secondaryValue = document.getElementById('secondary-field').value.trim();
                if (secondaryValue === '' || secondaryValue === document.getElementById('secondary-field').placeholder) {
                    showNotification('Please select a branch!', 'error');
                    hidePreloader();
                    return; // Prevent submission
                    }
                body = JSON.stringify({
                    [configField]: secondaryValue
                });
            } else {
                body = JSON.stringify({
                    [configField]: document.getElementById('primary-field').value
                });
            }
        } else if (step === "admin_account") { // User setup
            url = '/api/v2/auth/setup/register';
            body = JSON.stringify({
                username: document.getElementById('primary-field').value,
                password: document.getElementById('secondary-field').value
            });
        } else { // Login or changeuser
            url = mode === 'changeuser' ? '/api/v2/auth/adduser' : '/auth/login';
            body = JSON.stringify({
                username: document.getElementById('primary-field').value,
                password: document.getElementById('secondary-field').value
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
                method: 'POST',
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
                showNotification('Setup completed, Auth disabled!', 'success');
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
                const finalizeResponse = await fetch('/api/v2/auth/setup/finalize', {
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
                const response = await fetch('/api/v2/saveconfig', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ LanguageSetting: lang })
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