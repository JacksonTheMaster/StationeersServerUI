(() => {
    const worldIdInput = document.getElementById('WorldID');
    const difficultyInput = document.getElementById('Difficulty');
    const startConditionInput = document.getElementById('StartCondition');
    const startLocationInput = document.getElementById('StartLocation');
    const fillHintWrapper = document.getElementById('fill-hint-wraper');
    const fillHint = document.getElementById('fill-hint');

    if (!worldIdInput || !difficultyInput || !startConditionInput || !startLocationInput || !fillHintWrapper) {
        return;
    }

    let catalog = { worlds: [], difficulties: [] };

    function normalizedValue(input) {
        return input.value.trim();
    }

    function createCombobox(input) {
        const root = input.closest('.worldgen-combobox');
        const panel = document.getElementById(input.getAttribute('aria-controls'));
        const toggle = root?.querySelector('.worldgen-combobox-toggle');
        let options = [];
        let filteredOptions = [];
        let activeIndex = -1;

        function close() {
            root.classList.remove('is-open');
            panel.hidden = true;
            input.setAttribute('aria-expanded', 'false');
            input.removeAttribute('aria-activedescendant');
            activeIndex = -1;
        }

        function setActive(index) {
            const elements = [...panel.querySelectorAll('.worldgen-combobox-option')];
            if (!elements.length) {
                activeIndex = -1;
                return;
            }

            activeIndex = (index + elements.length) % elements.length;
            elements.forEach((element, optionIndex) => {
                const isActive = optionIndex === activeIndex;
                element.classList.toggle('is-active', isActive);
                element.setAttribute('aria-selected', String(isActive));
            });
            const active = elements[activeIndex];
            input.setAttribute('aria-activedescendant', active.id);
            active.scrollIntoView({ block: 'nearest' });
        }

        function choose(index) {
            const item = filteredOptions[index];
            if (!item) {
                return;
            }
            input.value = item.value;
            input.dispatchEvent(new Event('input', { bubbles: true }));
            input.dispatchEvent(new Event('change', { bubbles: true }));
            close();
            input.focus();
        }

        function render() {
            filteredOptions = options;
            panel.replaceChildren();

            filteredOptions.forEach((item, index) => {
                const option = document.createElement('button');
                const label = document.createElement('span');
                const value = document.createElement('span');
                option.type = 'button';
                option.id = `${panel.id}-option-${index}`;
                option.className = 'worldgen-combobox-option';
                option.classList.toggle('is-current', item.value === normalizedValue(input));
                option.setAttribute('role', 'option');
                option.setAttribute('aria-selected', 'false');
                option.dataset.value = item.value;
                label.className = 'worldgen-combobox-option-label';
                label.textContent = item.label || item.value;
                value.className = 'worldgen-combobox-option-value';
                value.textContent = item.label && item.label !== item.value ? item.value : '';
                option.append(label, value);
                option.addEventListener('pointerdown', event => event.preventDefault());
                option.addEventListener('click', () => choose(index));
                option.addEventListener('pointerenter', () => setActive(index));
                panel.appendChild(option);
            });

            activeIndex = -1;
            input.removeAttribute('aria-activedescendant');
            if (!filteredOptions.length) {
                close();
            }
        }

        function open() {
            if (input.disabled) {
                return;
            }
            render();
            if (!filteredOptions.length) {
                return;
            }
            document.querySelectorAll('.worldgen-combobox.is-open').forEach(element => {
                if (element !== root) {
                    element.dispatchEvent(new CustomEvent('worldgen-combobox-close'));
                }
            });
            root.classList.add('is-open');
            panel.hidden = false;
            input.setAttribute('aria-expanded', 'true');
        }

        function setOptions(nextOptions) {
            options = nextOptions.map(item => ({
                value: String(item.value ?? ''),
                label: String(item.label || item.value || '')
            })).filter(item => item.value);
            if (root.classList.contains('is-open')) {
                open();
            }
        }

        function syncDisabled() {
            toggle.disabled = input.disabled;
            root.classList.toggle('is-disabled', input.disabled);
            if (input.disabled) {
                close();
            }
        }

        input.addEventListener('focus', open);
        input.addEventListener('click', open);
        input.addEventListener('input', close);
        input.addEventListener('keydown', event => {
            if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
                event.preventDefault();
                if (!root.classList.contains('is-open')) {
                    open();
                }
                setActive(activeIndex + (event.key === 'ArrowDown' ? 1 : -1));
            } else if (event.key === 'Enter' && activeIndex >= 0) {
                event.preventDefault();
                choose(activeIndex);
            } else if (event.key === 'Escape') {
                event.preventDefault();
                close();
            } else if (event.key === 'Tab') {
                close();
            }
        });
        toggle.addEventListener('pointerdown', event => event.preventDefault());
        toggle.addEventListener('click', () => {
            if (root.classList.contains('is-open')) {
                close();
            } else {
                input.focus();
                open();
            }
        });
        root.addEventListener('worldgen-combobox-close', close);

        syncDisabled();
        return { close, setOptions, syncDisabled };
    }

    const comboboxes = {
        worlds: createCombobox(worldIdInput),
        difficulties: createCombobox(difficultyInput),
        conditions: createCombobox(startConditionInput),
        locations: createCombobox(startLocationInput)
    };

    document.addEventListener('pointerdown', event => {
        if (!event.target.closest('.worldgen-combobox')) {
            Object.values(comboboxes).forEach(combobox => combobox.close());
        }
    });

    function selectedWorld() {
        const worldId = normalizedValue(worldIdInput);
        return catalog.worlds.find(world => world.id === worldId);
    }

    function refreshWorldDependentOptions() {
        const world = selectedWorld();
        comboboxes.conditions.setOptions(world?.startConditions || []);
        comboboxes.locations.setOptions(world?.startLocations || []);
    }

    function setSequenceValidity() {
        const hasDifficulty = normalizedValue(difficultyInput) !== '';
        const hasCondition = normalizedValue(startConditionInput) !== '';
        const hasLocation = normalizedValue(startLocationInput) !== '';
        const conditionHasGap = hasCondition && !hasDifficulty;
        const locationHasGap = hasLocation && (!hasDifficulty || !hasCondition);
        const validationMessage = fillHint?.textContent.trim() || 'Fill world-generation values in order.';

        startConditionInput.setCustomValidity(conditionHasGap ? validationMessage : '');
        startLocationInput.setCustomValidity(locationHasGap ? validationMessage : '');
        startConditionInput.closest('.form-group')?.classList.toggle('worldgen-sequence-invalid', conditionHasGap);
        startLocationInput.closest('.form-group')?.classList.toggle('worldgen-sequence-invalid', locationHasGap);

        fillHintWrapper.hidden = !conditionHasGap && !locationHasGap;
        startConditionInput.disabled = !hasDifficulty && !hasCondition;
        startLocationInput.disabled = !hasCondition && !hasLocation;
        comboboxes.conditions.syncDisabled();
        comboboxes.locations.syncDisabled();
    }

    function clearAfterDifficulty() {
        if (normalizedValue(difficultyInput) === '') {
            startConditionInput.value = '';
            startLocationInput.value = '';
        }
        setSequenceValidity();
    }

    function clearAfterCondition() {
        if (normalizedValue(startConditionInput) === '') {
            startLocationInput.value = '';
        }
        setSequenceValidity();
    }

    worldIdInput.addEventListener('input', refreshWorldDependentOptions);
    worldIdInput.addEventListener('change', () => {
        startConditionInput.value = '';
        startLocationInput.value = '';
        refreshWorldDependentOptions();
        setSequenceValidity();
    });
    difficultyInput.addEventListener('input', setSequenceValidity);
    difficultyInput.addEventListener('change', clearAfterDifficulty);
    startConditionInput.addEventListener('input', setSequenceValidity);
    startConditionInput.addEventListener('change', clearAfterCondition);
    startLocationInput.addEventListener('input', setSequenceValidity);

    async function loadCatalog() {
        try {
            const response = await fetch('/api/v3/worldgen/catalog', {
                headers: { Accept: 'application/json' }
            });
            if (!response.ok) {
                throw new Error(`catalog request failed with ${response.status}`);
            }
            catalog = await response.json();
            comboboxes.worlds.setOptions(catalog.worlds.map(world => ({ value: world.id, label: world.label })));
            comboboxes.difficulties.setOptions(catalog.difficulties);
            refreshWorldDependentOptions();
        } catch (error) {
            console.warn('World-generation suggestions are unavailable; custom values remain usable.', error);
        }
    }

    setSequenceValidity();
    loadCatalog();
})();
