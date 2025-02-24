// 状態管理クラス
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
            console.log('WebSocket: 接続確立');
            this.showStatus('接続しました', 'success');
        };

        this.ws.onclose = () => {
            console.log('WebSocket: 接続切断');
            this.showStatus('接続が切断されました。再接続します...', 'error');
            setTimeout(() => this.connect(), 3000);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket エラー:', error);
            this.showStatus('エラーが発生しました: ' + error.message, 'error');
        };

        this.ws.onmessage = (event) => {
            console.log('WebSocket受信データ:', event.data);
            try {
                const data = JSON.parse(event.data);
                console.log('パース済みデータ:', data);
                this.handleStateUpdate(data);
            } catch (error) {
                console.error('データパースエラー:', error);
                this.showStatus('データ処理エラー: ' + error.message, 'error');
            }
        };
    }

    setupEventListeners() {
        // 状態遷移ボタンのイベントリスナー
        document.getElementById('activate-btn').addEventListener('click', () => {
            console.log('イベント送信: activate');
            this.controlAutoTransition('activate');
        });

        document.getElementById('next-btn').addEventListener('click', () => {
            console.log('イベント送信: next');
            this.controlAutoTransition('next');
        });

        document.getElementById('finish-btn').addEventListener('click', () => {
            console.log('イベント送信: finish');
            this.controlAutoTransition('finish');
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
            console.log('自動遷移開始リクエスト');
            this.controlAutoTransition('start');
        });

        document.getElementById('stop-auto').addEventListener('click', () => {
            console.log('自動遷移停止リクエスト');
            this.controlAutoTransition('stop');
        });

        document.getElementById('reset-btn').addEventListener('click', () => {
            console.log('リセットリクエスト');
            this.controlAutoTransition('reset');
        });
    }

    async controlAutoTransition(action) {
        console.log(`自動遷移API呼び出し: ${action}`);
        try {
            const response = await fetch(`/api/auto-transition?action=${action}`, {
                method: 'POST'
            });

            console.log('APIレスポンス:', {
                status: response.status,
                statusText: response.statusText
            });

            if (response.ok) {
                console.log(`自動遷移${action}成功`);
                this.showStatus(`自動遷移${action === 'start' ? '開始' : '停止'}`, 'success');
                this.updateAutoTransitionStatus(action === 'start');
            } else {
                const error = await response.text();
                console.error('自動遷移APIエラー:', error);
                this.showStatus(`自動遷移制御エラー: ${error}`, 'error');
            }
        } catch (error) {
            console.error('自動遷移API例外:', error);
            this.showStatus(`自動遷移制御エラー: ${error.message}`, 'error');
        }
    }

    updateAutoTransitionStatus(isRunning) {
        console.log('自動遷移状態更新:', isRunning);
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
        console.log('WebSocketイベント送信:', event);
        if (this.ws.readyState === WebSocket.OPEN) {
            const message = JSON.stringify({ event });
            console.log('送信メッセージ:', message);
            this.ws.send(message);
        } else {
            console.error('WebSocket接続エラー - 現在の状態:', this.ws.readyState);
            this.showStatus('サーバーに接続できません', 'error');
        }
    }

    handleStateUpdate(data) {
        console.log('状態更新データ受信:', data);
        if (data.type === 'error') {
            console.error('状態更新エラー:', data.error);
            this.showStatus(data.error, 'error');
            return;
        }

        console.log('新しい状態:', {
            state: data.state,
            phase: data.phase,
            nextTransition: data.next_transition,
            conditions: data.conditions
        });

        this.updateState(data);
        this.updateTransitionInfo(data);
        if (data.message) {
            console.log('状態メッセージ:', data.message);
            this.updateStateMessage(data.message);
        }
        this.updateConditions(data.conditions);
    }

    updateConditions(conditions) {
        console.log('条件更新:', conditions);
        const conditionsList = document.getElementById('conditions-list');
        conditionsList.innerHTML = '';

        conditions.forEach(condition => {
            const conditionElement = document.createElement('div');
            conditionElement.className = 'condition-item';
            
            const header = document.createElement('div');
            header.className = 'condition-header';
            
            const label = document.createElement('div');
            label.className = 'condition-label';
            label.textContent = `${condition.label} (Kind: ${condition.kind}, Clear: ${condition.is_clear})`;
            
            const state = document.createElement('div');
            state.className = `condition-state state-${condition.state}`;
            state.textContent = condition.state;
            
            const description = document.createElement('div');
            description.className = 'condition-description';
            description.textContent = condition.description || 'No description';
            
            header.appendChild(label);
            header.appendChild(state);
            conditionElement.appendChild(header);
            conditionElement.appendChild(description);

            if (condition.parts && condition.parts.length > 0) {
                const partsList = document.createElement('div');
                partsList.className = 'parts-list';
                
                condition.parts.forEach(part => {
                    const partElement = document.createElement('div');
                    partElement.className = 'part-item';
                    
                    const partInfo = document.createElement('div');
                    partInfo.className = 'part-info';
                    
                    const partBasic = document.createElement('div');
                    partBasic.className = 'part-basic';
                    partBasic.innerHTML = `
                        <strong>${part.label}</strong> (Clear: ${part.is_clear})<br>
                        State: <span class="state-${part.state}">${part.state}</span><br>
                        Operator: ${part.comparison_operator}
                    `;
                    
                    const partDetails = document.createElement('div');
                    partDetails.className = 'part-details';
                    partDetails.innerHTML = `
                        Target: ${part.target_entity_type} (ID: ${part.target_entity_id})<br>
                        Values: Int=${part.reference_value_int},
                               Float=${part.reference_value_float},
                               String="${part.reference_value_string}"<br>
                        Range: ${part.min_value} - ${part.max_value}<br>
                        Priority: ${part.priority}
                    `;
                    
                    partInfo.appendChild(partBasic);
                    partInfo.appendChild(partDetails);
                    partElement.appendChild(partInfo);
                    partsList.appendChild(partElement);
                });
                
                conditionElement.appendChild(partsList);
            }

            conditionsList.appendChild(conditionElement);
        });
    }

    formatTime(timeStr) {
        if (!timeStr) return '-';
        const date = new Date(timeStr);
        return date.toLocaleString('ja-JP', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    }

    updateState(data) {
        console.log('状態更新処理:', {
            currentState: this.currentState,
            newState: data.state,
            phase: data.phase
        });

        // 現在の状態表示を更新
        const currentStateElement = document.getElementById('current-state');
        currentStateElement.textContent = data.state;
        currentStateElement.className = `state-display ${data.phase?.is_clear ? 'is-clear' : 'not-clear'}`;
        this.currentState = data.state;

        // フェーズの詳細情報を更新
        if (data.phase) {
            document.getElementById('phase-name').textContent = data.phase.name || '-';
            document.getElementById('phase-description').textContent = data.phase.description || '-';
            document.getElementById('phase-order').textContent = data.phase.order || '-';
            document.getElementById('phase-start-time').textContent = this.formatTime(data.phase.start_time);
            document.getElementById('phase-finish-time').textContent = this.formatTime(data.phase.finish_time);
        }

        // 状態図の更新
        this.updateStateDiagram(data.state);

        // ボタンの有効/無効を更新
        this.updateButtons(data.state);

        // 完了状態の場合、自動遷移を停止
        if (data.state === 'finish') {
            console.log('完了状態検出 - 自動遷移停止');
            this.updateAutoTransitionStatus(false);
        }
    }

    updateStateMessage(message) {
        console.log('メッセージ更新:', message);
        const messageElement = document.getElementById('state-message');
        messageElement.textContent = message;
        messageElement.className = 'state-message ' + this.currentState;
    }

    updateTransitionInfo(data) {
        console.log('遷移情報更新:', {
            currentState: this.currentState,
            nextTransition: data.nextTransition
        });

        const nextTransitionElement = document.getElementById('next-transition');
        if (this.currentState === 'finish') {
            nextTransitionElement.textContent = '';
        } else {
            const nextTransition = new Date(data.nextTransition);
            this.startTransitionCountdown(nextTransition);
        }
    }

    startTransitionCountdown(nextTransition) {
        console.log('カウントダウン開始:', nextTransition);
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
        console.log('状態図更新:', newState);
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
        console.log('可能な遷移をハイライト:', state);
        const transitionMap = {
            'ready': ['activate'],
            'active': ['next'],
            'next': ['activate', 'finish']
        };

        const possibleTransitions = transitionMap[state] || [];
        console.log('可能な遷移:', possibleTransitions);
        possibleTransitions.forEach(event => {
            const transitionElement = document.querySelector(`.transition[data-event="${event}"]`);
            if (transitionElement) {
                transitionElement.classList.add('active');
            }
        });
    }

    updateButtons(state) {
        console.log('ボタン状態更新:', state);
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
        console.log('ステータス表示:', { message, type });
        const statusElement = document.getElementById('status-message');
        statusElement.textContent = message;
        statusElement.className = 'status-message ' + type;
    }
}

// アプリケーションの初期化
document.addEventListener('DOMContentLoaded', () => {
    console.log('アプリケーション初期化開始');
    new StateManager();
});