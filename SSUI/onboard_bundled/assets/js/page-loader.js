(() => {
    const root = document.documentElement;
    root.classList.add('ssui-loading');

    try {
        const savedTheme = JSON.parse(localStorage.getItem('ssui-theme'));
        if (savedTheme && typeof savedTheme === 'object') {
            Object.entries(savedTheme).forEach(([name, value]) => {
                if (name.startsWith('--') && value) root.style.setProperty(name, value);
            });
        }
    } catch {
        // A broken saved theme should never keep the interface hidden.
    }

    const pageLoaded = new Promise(resolve => {
        if (document.readyState === 'complete') resolve();
        else window.addEventListener('load', resolve, { once: true });
    });
    let revealing = false;

    async function revealPage() {
        if (revealing) return;
        revealing = true;
        const backgroundReady = window.SSUIBackground?.ready || Promise.resolve();
        const timeout = new Promise(resolve => window.setTimeout(resolve, 10000));
        await Promise.race([Promise.all([pageLoaded, backgroundReady]), timeout]);
        await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));

        const loader = document.getElementById('page-loader');
        root.classList.remove('ssui-loading');
        if (!loader) return;
        loader.classList.add('page-loader-leaving');
        window.setTimeout(() => loader.remove(), 200);
    }

    document.addEventListener('DOMContentLoaded', revealPage, { once: true });
})();
