// WebSocket接続の管理
class StateManager {
    constructor() {
        this.connect();
        this.setupEventListeners();
        this.currentState = 'ready';
        this.setupAutoTransitionControls();
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        this.ws = new WebSocket(`${protocol}//${window.location.host}/ws`);

        this.ws.onopen = () => {
            this.showStatus('接続しました', 'success');
        };

        this.ws.onclose = () => {
            this.showStatus('接続が切断されました。再接続します...', 'error');
            setTimeout(() => this.connect(), 3000);
        };

        this.ws.onerror = (error) => {
            this.showStatus('エラーが発生しました: ' + error.message, 'error');
        };

        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.handleStateUpdate(data);
        };
    }

    setupEventListeners() {
        // 状態遷移ボタンのイベントリスナー
        document.getElementById('activate-btn').addEventListener('click', () => {
            this.sendEvent('activate');
        });

        document.getElementById('next-btn').addEventListener('click', () => {
            this.sendEvent('next');
        });

        document.getElementById('finish-btn').addEventListener('click', () => {
            this.sendEvent('finish');
        });
    }

    setupAutoTransitionControls() {
        // 自動遷移制御ボタンの追加
        const controlsDiv = document.querySelector('.controls');
        const autoTransitionDiv = document.createElement('div');
        autoTransitionDiv.className = 'auto-transition-controls';
        autoTransitionDiv.innerHTML = `
            <button id="start-auto" class="control-btn">自動遷移開始</button>
            <button id="stop-auto" class="control-btn" disabled>自動遷移停止</button>
            <button id="reset-btn" class="control-btn">リセット</button>
            <div id="next-transition" class="transition-info"></div>
            <div id="state-message" class="state-message"></div>
        `;
        controlsDiv.appendChild(autoTransitionDiv);

        // 自動遷移ボタンのイベントリスナー
        document.getElementById('start-auto').addEventListener('click', () => {
            this.controlAutoTransition('start');
        });

        document.getElementById('stop-auto').addEventListener('click', () => {
            this.controlAutoTransition('stop');
        });

        document.getElementById('reset-btn').addEventListener('click', () => {
            this.sendEvent('reset');
        });
    }

    async controlAutoTransition(action) {
        try {
            const response = await fetch(`/api/auto-transition?action=${action}`, {
                method: 'POST'
            });

            if (response.ok) {
                this.showStatus(`自動遷移${action === 'start' ? '開始' : '停止'}`, 'success');
                this.updateAutoTransitionStatus(action === 'start');
            } else {
                const error = await response.text();
                this.showStatus(`自動遷移制御エラー: ${error}`, 'error');
            }
        } catch (error) {
            this.showStatus(`自動遷移制御エラー: ${error.message}`, 'error');
        }
    }

    updateAutoTransitionStatus(isRunning) {
        const startBtn = document.getElementById('start-auto');
        const stopBtn = document.getElementById('stop-auto');
        const resetBtn = document.getElementById('reset-btn');
        
        if (isRunning) {
            startBtn.disabled = true;
            stopBtn.disabled = false;
            resetBtn.disabled = true;
        } else {
            startBtn.disabled = false;
            stopBtn.disabled = true;
            resetBtn.disabled = false;
        }
    }

    sendEvent(event) {
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ event }));
        } else {
            this.showStatus('サーバーに接続できません', 'error');
        }
    }

    handleStateUpdate(data) {
        if (data.type === 'error') {
            this.showStatus(data.error, 'error');
            return;
        }

        if (data.type === 'state_change') {
            this.updateState(data);
            this.updateTransitionInfo(data);
            if (data.info && data.info.message) {
                this.updateStateMessage(data.info.message);
            }
        }
    }

    updateState(data) {
        // 現在の状態表示を更新
        const currentStateElement = document.getElementById('current-state');
        currentStateElement.textContent = data.state;
        this.currentState = data.state;

        // 状態図の更新
        this.updateStateDiagram(data.state);

        // ボタンの有効/無効を更新
        this.updateButtons(data.state);

        // 完了状態の場合、自動遷移を停止
        if (data.state === 'finish') {
            this.updateAutoTransitionStatus(false);
        }
    }

    updateStateMessage(message) {
        const messageElement = document.getElementById('state-message');
        messageElement.textContent = message;
        messageElement.className = 'state-message ' + this.currentState;
    }

    updateTransitionInfo(data) {
        const nextTransitionElement = document.getElementById('next-transition');
        if (data.nextTransition && this.currentState !== 'finish') {
            const nextTransition = new Date(data.nextTransition);
            this.startTransitionCountdown(nextTransition);
        } else {
            nextTransitionElement.textContent = '';
        }
    }

    startTransitionCountdown(nextTransition) {
        const nextTransitionElement = document.getElementById('next-transition');
        
        // 既存のカウントダウンをクリア
        if (this.countdownInterval) {
            clearInterval(this.countdownInterval);
        }

        const updateCountdown = () => {
            const now = new Date();
            const timeLeft = Math.max(0, (nextTransition - now) / 1000);
            
            if (timeLeft > 0) {
                nextTransitionElement.textContent = `次の遷移まで: ${Math.ceil(timeLeft)}秒`;
            } else {
                nextTransitionElement.textContent = '遷移待機中...';
                clearInterval(this.countdownInterval);
            }
        };

        updateCountdown();
        this.countdownInterval = setInterval(updateCountdown, 1000);
    }

    updateStateDiagram(newState) {
        // 全ての状態をリセット
        document.querySelectorAll('.state').forEach(state => {
            state.classList.remove('active');
        });

        // 全ての遷移をリセット
        document.querySelectorAll('.transition').forEach(transition => {
            transition.classList.remove('active');
        });

        // 現在の状態をアクティブに
        const currentStateElement = document.querySelector(`.state[data-state="${newState}"]`);
        if (currentStateElement) {
            currentStateElement.classList.add('active');
        }

        // 可能な遷移をハイライト
        this.highlightPossibleTransitions(newState);
    }

    highlightPossibleTransitions(state) {
        const transitionMap = {
            'ready': ['activate'],
            'active': ['next'],
            'next': ['activate', 'finish']
        };

        const possibleTransitions = transitionMap[state] || [];
        possibleTransitions.forEach(event => {
            const transitionElement = document.querySelector(`.transition[data-event="${event}"]`);
            if (transitionElement) {
                transitionElement.classList.add('active');
            }
        });
    }

    updateButtons(state) {
        // ボタンの有効/無効を状態に応じて更新
        const buttons = {
            'activate-btn': state === 'ready',
            'next-btn': state === 'active',
            'finish-btn': state === 'next'
        };

        Object.entries(buttons).forEach(([id, enabled]) => {
            const button = document.getElementById(id);
            button.disabled = !enabled;
        });
    }

    showStatus(message, type) {
        const statusElement = document.getElementById('status-message');
        statusElement.textContent = message;
        statusElement.className = 'status-message ' + type;
    }
}

// アプリケーションの初期化
document.addEventListener('DOMContentLoaded', () => {
    new StateManager();
});