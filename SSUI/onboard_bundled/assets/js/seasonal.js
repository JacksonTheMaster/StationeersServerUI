(() => {
    const seasons = [
        {
            id: 'halloween',
            from: 1020,
            through: 1102,
            decorations: ['🦇', '🦇', '🎃', '👻']
        },
        {
            id: 'christmas',
            from: 1201,
            through: 110,
            banner: '/static/seasonal/christmas/stationeers-winter.webp',
            decorations: [
                '/static/seasonal/christmas/c1.webp',
                '/static/seasonal/christmas/c2.webp',
                '/static/seasonal/christmas/c3.webp',
                '/static/seasonal/christmas/c4.webp'
            ]
        }
    ];

    function today(date) {
        return (date.getMonth() + 1) * 100 + date.getDate();
    }

    function isActive(season, date) {
        const value = today(date);
        if (season.from <= season.through) {
            return value >= season.from && value <= season.through;
        }
        return value >= season.from || value <= season.through;
    }

    function addDecorations(season) {
        const layer = document.createElement('div');
        layer.id = 'seasonal-layer';
        layer.className = `seasonal-layer seasonal-layer--${season.id}`;
        layer.setAttribute('aria-hidden', 'true');

        season.decorations.forEach((decoration, index) => {
            const item = decoration.startsWith('/') ? document.createElement('img') : document.createElement('span');
            item.className = `seasonal-decoration seasonal-decoration--${index + 1}`;
            if (item instanceof HTMLImageElement) {
                item.src = decoration;
                item.alt = '';
                item.draggable = false;
            } else {
                item.textContent = decoration;
            }
            layer.appendChild(item);
        });

        document.body.appendChild(layer);
    }

    function applySeason(season) {
        document.body.classList.add('seasonal-active', `seasonal-${season.id}`);
        document.documentElement.dataset.season = season.id;
        addDecorations(season);

        if (season.banner) {
            const banner = document.getElementById('banner');
            if (banner) banner.src = season.banner;
        }

        window.addEventListener('focus', () => document.body.classList.remove('seasonal-paused'));
        window.addEventListener('blur', () => document.body.classList.add('seasonal-paused'));
        document.addEventListener('visibilitychange', () => {
            document.body.classList.toggle('seasonal-paused', document.hidden);
        });
    }

    const preview = new URLSearchParams(window.location.search).get('season');
    const active = seasons.find(season => season.id === preview) || seasons.find(season => isActive(season, new Date()));
    window.SSUISeasonal = { active: active?.id || null };
    if (active) applySeason(active);
})();
