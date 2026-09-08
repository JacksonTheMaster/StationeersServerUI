(() => {
    const storageKey = 'ssuiBackgroundModeV2';
    const sceneKey = 'ssuiPanoramaScene';
    const bodyKey = 'ssuiCelestialBody';
    const panoramaCycleKey = 'ssuiPanoramaCycleStart';
    const celestialCycleKey = 'ssuiCelestialCycleStart';
    const modes = {
        'celestial-focus': 'Celestial view',
        'celestial-static': 'Celestial view (reduced motion)',
        'panorama-focus': 'Station panoramas',
        'panorama-static': 'Station panoramas (reduced motion)',
        'planets-focus': 'Classic planets',
        'planets-static': 'Classic planets (reduced motion)',
        'disabled': 'No background'
    };

    const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const hardwareWebGL = detectHardwareWebGL();
    let mode = localStorage.getItem(storageKey);
    let frame;
    let renderer;
    let picker;
    let rendererGeneration = 0;
    let windowFocused = true;
    let resolveReady;
    let readyResolved = false;
    const ready = new Promise(resolve => resolveReady = resolve);
    if (!modes[mode]) mode = reducedMotion || !hardwareWebGL ? 'celestial-static' : 'celestial-focus';

    function detectHardwareWebGL() {
        const canvas = document.createElement('canvas');
        const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
        if (!gl) return false;

        const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
        const renderer = debugInfo ? gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL) : '';
        gl.getExtension('WEBGL_lose_context')?.loseContext();
        return !/swiftshader|llvmpipe|softpipe|software rasterizer|software renderer/i.test(renderer);
    }

    function baseMode() {
        return mode.replace(/-(?:focus|static)$/, '');
    }

    function shouldAnimate() {
        return !reducedMotion && !mode.endsWith('-static');
    }

    function isPaused() {
        return mode !== 'disabled' && (!windowFocused || !shouldAnimate());
    }

    function setBodyMode() {
        document.body.classList.remove('background-celestial', 'background-panorama', 'background-planets', 'background-disabled', 'background-paused');
        document.body.classList.add(baseMode() === 'celestial' ? 'background-celestial' :
            baseMode() === 'panorama' ? 'background-panorama' :
                baseMode() === 'planets' ? 'background-planets' : 'background-disabled');
        document.body.classList.toggle('background-paused', isPaused());
    }

    function createPlanets() {
        let container = document.getElementById('planet-container');
        if (!container) {
            container = document.createElement('div');
            container.id = 'planet-container';
            document.body.prepend(container);
        }
        if (container.childElementCount) return;

        addPlanet(container, 80, 650, 34, 'rgba(200, 100, 50, 0.7)');
        addPlanet(container, 50, 1000, 46, 'rgba(100, 200, 150, 0.5)');
        addPlanet(container, 30, 1250, 63, 'rgba(50, 150, 250, 0.6)');
        addPlanet(container, 70, 400, 28, 'rgba(200, 150, 200, 0.7)');
    }

    function addPlanet(container, size, orbitRadius, speed, color) {
        const orbit = document.createElement('div');
        orbit.className = 'orbit';
        orbit.style.cssText = `width:${orbitRadius * 2}px;height:${orbitRadius * 2}px;position:absolute;left:50%;top:50%;animation:orbit ${speed}s linear infinite ${-(Math.random() * speed)}s`;

        const planet = document.createElement('div');
        planet.className = 'planet';
        planet.style.cssText = `width:${size}px;height:${size}px;position:absolute;left:0;top:50%;background:${color};border-radius:50%;box-shadow:0 0 20px ${color}`;
        orbit.appendChild(planet);
        container.appendChild(orbit);
    }

    function updateButton() {
        const button = document.querySelector('.gpusaver-icon');
        if (!button) return;
        button.title = `Background: ${modes[mode]}`;
        button.setAttribute('aria-label', `Background: ${modes[mode]}`);
        button.dataset.backgroundMode = mode;
    }

    function buildPicker() {
        const button = document.querySelector('.gpusaver-icon');
        if (!button || picker) return;

        picker = document.createElement('div');
        picker.className = 'background-picker';
        picker.hidden = true;
        picker.innerHTML = `
            <strong>Background</strong>
            ${Object.entries(modes).map(([value, label]) => `
                <label><input type="radio" name="ssui-background-mode" value="${value}">
                    <span>${label}</span></label>`).join('')}`;
        button.insertAdjacentElement('afterend', picker);

        picker.addEventListener('change', event => {
            if (event.target.name === 'ssui-background-mode') setMode(event.target.value);
        });
        document.addEventListener('click', event => {
            if (!picker.hidden && !picker.contains(event.target) && event.target !== button) picker.hidden = true;
        });
        document.addEventListener('keydown', event => {
            if (event.key === 'Escape') picker.hidden = true;
        });
    }

    function openPicker() {
        buildPicker();
        if (!picker) return;
        const selected = picker.querySelector(`input[value="${mode}"]`);
        if (selected) selected.checked = true;
        picker.hidden = !picker.hidden;
    }

    function setMode(nextMode) {
        if (!modes[nextMode]) return;
        mode = nextMode;
        localStorage.setItem(storageKey, mode);
        setBodyMode();
        updateButton();
        if (picker) picker.hidden = true;

        if (baseMode() === 'planets') createPlanets();
        stopRenderer();
        if (baseMode() === 'celestial') startCelestial();
        if (baseMode() === 'panorama') startPanorama();

        window.SSUIAccess?.notify(`Background set to ${modes[mode]}.`, 'success');
    }

    function stopRenderer() {
        rendererGeneration++;
        window.cancelAnimationFrame(frame);
        frame = null;
        disposeRenderer(renderer);
        renderer = null;
    }

    function disposeRenderer(value) {
        if (!value) return;
        value.gl.getExtension('WEBGL_lose_context')?.loseContext();
        value.canvas.remove();
    }

    async function startPanorama() {
        if (renderer) {
            resumeAnimation();
            return;
        }

        const generation = ++rendererGeneration;
        try {
            const manifestResponse = await fetch('/static/backgrounds/panoramas/manifest.json');
            if (!manifestResponse.ok) throw new Error('manifest unavailable');
            const manifest = await manifestResponse.json();
            if (!manifest.scenes?.length) throw new Error('no panorama scenes installed');
            const scene = chooseItem(manifest.scenes, sceneKey);
            const created = await createPanoramaRenderer(scene);
            if (generation !== rendererGeneration || baseMode() !== 'panorama') {
                disposeRenderer(created);
                return;
            }
            renderer = created;
            renderer.canvas.dataset.scene = scene.id;
            document.getElementById('space-background').prepend(renderer.canvas);
            document.body.classList.remove('background-fallback');
            drawBackground(performance.now());
            resumeAnimation();
            preloadPanoramas(manifest.scenes, scene);
            preloadCelestialPool();
        } catch (error) {
            document.body.classList.add('background-fallback');
            console.info(`SSUI panorama background unavailable: ${error.message}`);
        }
    }

    async function startCelestial(configuredWorldID) {
        if (renderer) {
            resumeAnimation();
            return;
        }

        const generation = ++rendererGeneration;
        try {
            const manifestResponse = await fetch('/static/backgrounds/celestial/manifest.json');
            if (!manifestResponse.ok) throw new Error('manifest unavailable');
            const manifest = await manifestResponse.json();
            if (!manifest.bodies?.length) throw new Error('no celestial bodies installed');
            const body = await chooseCelestialBody(manifest.bodies, configuredWorldID);
            const created = await createCelestialRenderer(body);
            if (generation !== rendererGeneration || baseMode() !== 'celestial') {
                disposeRenderer(created);
                return;
            }
            renderer = created;
            renderer.canvas.dataset.scene = body.id;
            document.getElementById('space-background').prepend(renderer.canvas);
            document.body.classList.remove('background-fallback');
            drawBackground(performance.now());
            resumeAnimation();
            preloadCelestialBodies(manifest.bodies, body);
            preloadPanoramaPool();
        } catch (error) {
            document.body.classList.add('background-fallback');
            console.info(`SSUI celestial background unavailable: ${error.message}`);
        }
    }

    function chooseItem(items, key) {
        const previous = sessionStorage.getItem(key);
        const choices = items.length > 1 ? items.filter(item => item.id !== previous) : items;
        const selected = choices[Math.floor(Math.random() * choices.length)];
        sessionStorage.setItem(key, selected.id);
        return selected;
    }

    function celestialAssetID(worldID) {
        const value = String(worldID || 'Lunar').toLowerCase();
        if (value.startsWith('mimas')) return 'mimas';
        if (value.startsWith('europa')) return 'europa';
        if (value.startsWith('mars')) return 'mars';
        if (value.startsWith('venus')) return 'venus';
        if (value.startsWith('vulcan')) return 'europa';
        if (value === 'lunar' || value === 'moon') return 'moon';
        return 'moon';
    }

    async function chooseCelestialBody(items, configuredWorldID) {
        let worldID = configuredWorldID || 'Lunar';
        try {
            if (!configuredWorldID) {
                const response = await fetch('/api/v3/server/status');
                if (response.ok) worldID = (await response.json()).worldId || worldID;
            }
        } catch (_) {}

        const wanted = celestialAssetID(worldID);
        return items.find(item => item.id === wanted) || items.find(item => item.id === 'moon') || items[0];
    }

    async function refreshCelestialBody(worldID) {
        if (baseMode() !== 'celestial') return;
        const space = document.getElementById('space-background');
        if (space) space.classList.add('background-switching');
        stopRenderer();
        await startCelestial(worldID);
        if (space) window.setTimeout(() => space.classList.remove('background-switching'), 40);
    }

    async function createPanoramaRenderer(scene) {
        const canvas = document.createElement('canvas');
        canvas.id = 'panorama-background';
        canvas.setAttribute('aria-hidden', 'true');
        const gl = canvas.getContext('webgl', { alpha: false, antialias: false, powerPreference: 'low-power' });
        if (!gl) throw new Error('WebGL is not available');

        const program = createProgram(gl);
        const texture = gl.createTexture();
        gl.bindTexture(gl.TEXTURE_CUBE_MAP, texture);
        gl.texParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
        gl.texParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
        gl.texParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
        gl.texParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);

        const root = scene.path || `/static/backgrounds/panoramas/${scene.id}`;
        const faces = [
            [gl.TEXTURE_CUBE_MAP_POSITIVE_X, 'right'],
            [gl.TEXTURE_CUBE_MAP_NEGATIVE_X, 'left'],
            [gl.TEXTURE_CUBE_MAP_POSITIVE_Y, 'up'],
            [gl.TEXTURE_CUBE_MAP_NEGATIVE_Y, 'down'],
            [gl.TEXTURE_CUBE_MAP_POSITIVE_Z, 'front'],
            [gl.TEXTURE_CUBE_MAP_NEGATIVE_Z, 'back']
        ];
        const images = await Promise.all(faces.map(([, name]) => loadImage(`${root}/${name}.webp`)));
        faces.forEach(([target], index) => gl.texImage2D(target, 0, gl.RGB, gl.RGB, gl.UNSIGNED_BYTE, images[index]));

        bindFullScreenTriangle(gl, program);

        const storedCycleStart = Number(sessionStorage.getItem(panoramaCycleKey));
        const cycleStart = Number.isFinite(storedCycleStart) && storedCycleStart > 0 ? storedCycleStart : Date.now();
        sessionStorage.setItem(panoramaCycleKey, String(cycleStart));

        return {
            kind: 'cube',
            canvas,
            gl,
            program,
            yaw: Math.random() * Math.PI * 2,
            startedAt: performance.now() - (Date.now() - cycleStart),
            aspect: gl.getUniformLocation(program, 'aspect'),
            pitch: gl.getUniformLocation(program, 'pitch'),
            yawLocation: gl.getUniformLocation(program, 'yaw')
        };
    }

    async function createCelestialRenderer(body) {
        const canvas = document.createElement('canvas');
        canvas.id = 'panorama-background';
        canvas.setAttribute('aria-hidden', 'true');
        const gl = canvas.getContext('webgl', { alpha: false, antialias: false, powerPreference: 'low-power' });
        if (!gl) throw new Error('WebGL is not available');

        const program = createCelestialProgram(gl);
        const image = await loadImage(body.path || `/static/backgrounds/celestial/${body.id}.webp`);
        const texture = gl.createTexture();
        gl.activeTexture(gl.TEXTURE0);
        gl.bindTexture(gl.TEXTURE_2D, texture);
        gl.pixelStorei(gl.UNPACK_FLIP_Y_WEBGL, true);
        gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, image);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
        createStarTexture(gl);
        bindFullScreenTriangle(gl, program);

        const storedCycleStart = Number(sessionStorage.getItem(celestialCycleKey));
        const cycleStart = Number.isFinite(storedCycleStart) && storedCycleStart > 0 ? storedCycleStart : Date.now();
        sessionStorage.setItem(celestialCycleKey, String(cycleStart));

        return {
            kind: 'celestial',
            canvas,
            gl,
            program,
            seed: Math.random() * 20,
            startedAt: performance.now() - (Date.now() - cycleStart),
            aspect: gl.getUniformLocation(program, 'aspect'),
            time: gl.getUniformLocation(program, 'time'),
            seedLocation: gl.getUniformLocation(program, 'seed')
        };
    }

    function bindFullScreenTriangle(gl, program) {
        const position = gl.getAttribLocation(program, 'position');
        const buffer = gl.createBuffer();
        gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
        gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 3, -1, -1, 3]), gl.STATIC_DRAW);
        gl.enableVertexAttribArray(position);
        gl.vertexAttribPointer(position, 2, gl.FLOAT, false, 0, 0);
    }

    function loadImage(url) {
        return new Promise((resolve, reject) => {
            const image = new Image();
            image.decoding = 'async';
            image.onload = () => resolve(image);
            image.onerror = () => reject(new Error(`could not load ${url}`));
            image.src = url;
        });
    }

    function preloadPanoramas(scenes, selected) {
        const preload = async () => {
            for (const scene of scenes) {
                if (scene.id === selected.id) continue;
                const root = scene.path || `/static/backgrounds/panoramas/${scene.id}`;
                const names = ['right', 'left', 'up', 'down', 'front', 'back'];
                await Promise.all(names.map(name => loadImage(`${root}/${name}.webp`))).catch(() => {});
            }
        };

        if ('requestIdleCallback' in window) window.requestIdleCallback(preload, { timeout: 2500 });
        else window.setTimeout(preload, 500);
    }

    function preloadCelestialBodies(bodies, selected) {
        const preload = async () => {
            for (const body of bodies) {
                if (body.id === selected.id) continue;
                await loadImage(body.path || `/static/backgrounds/celestial/${body.id}.webp`).catch(() => {});
            }
        };

        if ('requestIdleCallback' in window) window.requestIdleCallback(preload, { timeout: 2500 });
        else window.setTimeout(preload, 500);
    }

    function preloadPanoramaPool() {
        fetch('/static/backgrounds/panoramas/manifest.json')
            .then(response => response.ok ? response.json() : Promise.reject())
            .then(manifest => preloadPanoramas(manifest.scenes || [], { id: '' }))
            .catch(() => {});
    }

    function preloadCelestialPool() {
        fetch('/static/backgrounds/celestial/manifest.json')
            .then(response => response.ok ? response.json() : Promise.reject())
            .then(manifest => preloadCelestialBodies(manifest.bodies || [], { id: '' }))
            .catch(() => {});
    }

    function createStarTexture(gl) {
        const canvas = document.createElement('canvas');
        canvas.width = 1024;
        canvas.height = 1024;
        const context = canvas.getContext('2d');
        context.fillStyle = '#000';
        context.fillRect(0, 0, canvas.width, canvas.height);

        let randomState = 0x51f15e;
        const random = () => {
            randomState = (randomState * 1664525 + 1013904223) >>> 0;
            return randomState / 4294967296;
        };
        for (let index = 0; index < 1150; index++) {
            const x = random() * canvas.width;
            const y = random() * canvas.height;
            const bright = random() > 0.91;
            const radius = bright ? 0.9 + random() * 1.25 : 0.28 + random() * 0.55;
            const warmth = random();
            context.globalAlpha = bright ? 0.65 + random() * 0.35 : 0.28 + random() * 0.48;
            context.fillStyle = warmth > 0.82 ? '#ffd8ad' : warmth < 0.18 ? '#a9c8ff' : '#f2f5ff';
            context.beginPath();
            context.arc(x, y, radius, 0, Math.PI * 2);
            context.fill();
        }
        context.globalAlpha = 1;

        const texture = gl.createTexture();
        gl.activeTexture(gl.TEXTURE1);
        gl.bindTexture(gl.TEXTURE_2D, texture);
        gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGB, gl.RGB, gl.UNSIGNED_BYTE, canvas);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT);
        gl.activeTexture(gl.TEXTURE0);
    }

    function createProgram(gl) {
        const vertex = compileShader(gl, gl.VERTEX_SHADER, `
            attribute vec2 position;
            varying vec2 uv;
            void main() {
                uv = position;
                gl_Position = vec4(position, 0.0, 1.0);
            }
        `);
        const fragment = compileShader(gl, gl.FRAGMENT_SHADER, `
            precision mediump float;
            varying vec2 uv;
            uniform samplerCube panorama;
            uniform float aspect;
            uniform float yaw;
            uniform float pitch;
            void main() {
                vec3 ray = normalize(vec3(uv.x * aspect, uv.y, 1.65));
                float cy = cos(yaw), sy = sin(yaw);
                float cp = cos(pitch), sp = sin(pitch);
                ray = vec3(cy * ray.x + sy * ray.z, ray.y, -sy * ray.x + cy * ray.z);
                ray = vec3(ray.x, cp * ray.y - sp * ray.z, sp * ray.y + cp * ray.z);
                gl_FragColor = textureCube(panorama, ray);
            }
        `);
        const program = gl.createProgram();
        gl.attachShader(program, vertex);
        gl.attachShader(program, fragment);
        gl.linkProgram(program);
        if (!gl.getProgramParameter(program, gl.LINK_STATUS)) throw new Error(gl.getProgramInfoLog(program));
        gl.useProgram(program);
        gl.uniform1i(gl.getUniformLocation(program, 'panorama'), 0);
        return program;
    }

    function createCelestialProgram(gl) {
        const vertex = compileShader(gl, gl.VERTEX_SHADER, `
            attribute vec2 position;
            varying vec2 uv;
            void main() {
                uv = position;
                gl_Position = vec4(position, 0.0, 1.0);
            }
        `);
        const fragment = compileShader(gl, gl.FRAGMENT_SHADER, `
            precision mediump float;
            varying vec2 uv;
            uniform sampler2D celestialBody;
            uniform sampler2D starField;
            uniform float aspect;
            uniform float time;
            uniform float seed;

            void main() {
                vec2 space = vec2(uv.x * aspect, uv.y);
                vec3 color = mix(vec3(0.002, 0.004, 0.012), vec3(0.008, 0.014, 0.035), uv.y * 0.5 + 0.5);
                vec2 farStars = fract(space * 0.16 + vec2(seed * 0.07 + time * 0.00035, seed * 0.11));
                vec2 nearStars = fract(space * 0.31 + vec2(seed * 0.13 - time * 0.0008, seed * 0.19 + time * 0.0002));
                color += texture2D(starField, farStars).rgb * 0.72;
                color += texture2D(starField, nearStars).rgb * 0.42;
                float haze = max(0.0, 1.0 - length(space - vec2(sin(seed), cos(seed)) * 0.75) * 0.58);
                color += vec3(0.018, 0.025, 0.055) * haze * haze;

                float travel = mod(time / 300.0, 1.0);
                float bodySize = mix(1.38, 0.20, travel);
                vec2 bodyCenter = vec2(
                    mix(-0.75, 0.72, travel),
                    mix(0.02, 0.52, travel)
                );
                vec2 bodyPosition = vec2(
                    (space.x - bodyCenter.x) / bodySize,
                    (space.y - bodyCenter.y) / bodySize
                );
                vec2 bodyUv = bodyPosition * 0.5 + 0.5;
                if (bodyUv.x >= 0.0 && bodyUv.x <= 1.0 && bodyUv.y >= 0.0 && bodyUv.y <= 1.0) {
                    vec4 body = texture2D(celestialBody, bodyUv);
                    color = mix(color, body.rgb, body.a);
                }

                gl_FragColor = vec4(color, 1.0);
            }
        `);
        const program = gl.createProgram();
        gl.attachShader(program, vertex);
        gl.attachShader(program, fragment);
        gl.linkProgram(program);
        if (!gl.getProgramParameter(program, gl.LINK_STATUS)) throw new Error(gl.getProgramInfoLog(program));
        gl.useProgram(program);
        gl.uniform1i(gl.getUniformLocation(program, 'celestialBody'), 0);
        gl.uniform1i(gl.getUniformLocation(program, 'starField'), 1);
        return program;
    }

    function compileShader(gl, type, source) {
        const shader = gl.createShader(type);
        gl.shaderSource(shader, source);
        gl.compileShader(shader);
        if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) throw new Error(gl.getShaderInfoLog(shader));
        return shader;
    }

    function resizeCanvas() {
        if (!renderer) return;
        const ratio = Math.min(window.devicePixelRatio || 1, 1.5);
        const width = Math.round(window.innerWidth * ratio);
        const height = Math.round(window.innerHeight * ratio);
        if (renderer.canvas.width === width && renderer.canvas.height === height) return;
        renderer.canvas.width = width;
        renderer.canvas.height = height;
        renderer.gl.viewport(0, 0, width, height);
    }

    function drawBackground(now) {
        resizeCanvas();
        const elapsed = now - renderer.startedAt;
        const gl = renderer.gl;
        gl.useProgram(renderer.program);
        gl.uniform1f(renderer.aspect, renderer.canvas.width / renderer.canvas.height);
        if (renderer.kind === 'celestial') {
            gl.uniform1f(renderer.time, elapsed * 0.001);
            gl.uniform1f(renderer.seedLocation, renderer.seed);
            gl.drawArrays(gl.TRIANGLES, 0, 3);
            return;
        }
        gl.uniform1f(renderer.yawLocation, renderer.yaw + elapsed * 0.000012);
        gl.uniform1f(renderer.pitch, Math.sin(elapsed * 0.00008) * 0.025);
        gl.drawArrays(gl.TRIANGLES, 0, 3);
    }

    function renderBackground(now) {
        if (!renderer || (baseMode() !== 'celestial' && baseMode() !== 'panorama')) return;
        if (isPaused()) {
            frame = null;
            return;
        }

        drawBackground(now);
        frame = window.requestAnimationFrame(renderBackground);
    }

    function resumeAnimation() {
        setBodyMode();

        if (isPaused()) {
            window.cancelAnimationFrame(frame);
            frame = null;
            return;
        }

        if (renderer && reducedMotion) {
            window.cancelAnimationFrame(frame);
            frame = null;
            drawBackground(renderer.startedAt);
            return;
        }
        if (renderer && !frame && !isPaused()) frame = window.requestAnimationFrame(renderBackground);
    }

    function focusChanged(focused) {
        windowFocused = focused;
        resumeAnimation();
    }

    async function initialize() {
        // Register these before loading textures so a blur during startup is
        // not missed.
        window.addEventListener('focus', () => focusChanged(true));
        window.addEventListener('blur', () => focusChanged(false));
        document.addEventListener('visibilitychange', () => focusChanged(!document.hidden && document.hasFocus()));

        setBodyMode();
        buildPicker();
        updateButton();
        if (baseMode() === 'planets') createPlanets();
        if (baseMode() === 'celestial') await startCelestial();
        if (baseMode() === 'panorama') await startPanorama();

        window.addEventListener('resize', resizeCanvas);
        document.querySelectorAll('[name="WorldID"]').forEach(input => {
            input.addEventListener('input', () => refreshCelestialBody(input.value));
        });

        if (!readyResolved) {
            readyResolved = true;
            resolveReady();
        }
    }

    window.toggleGPUSaver = openPicker;
    window.SSUIBackground = { openPicker, setMode, ready };
    if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', initialize);
    else initialize();
})();
