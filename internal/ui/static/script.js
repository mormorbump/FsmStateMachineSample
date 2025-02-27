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

    async handleCounterIncrement(conditionId, partId, increment = 1) {
        try {
            const response = await fetch(`/api/condition/${conditionId}/part/${partId}/evaluate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ increment })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            this.showStatus(`カウンター更新: ${result.current_value}`, 'success');
            return result;
        } catch (error) {
            console.error('カウンター更新エラー:', error);
            this.showStatus(`カウンター更新エラー: ${error.message}`, 'error');
            throw error;
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

                    // 基本情報の表示
                    partBasic.innerHTML = `
                        <strong>${part.label}</strong> (Clear: ${part.is_clear})<br>
                        State: <span class="state-${part.state}">${part.state}</span><br>
                        Operator: ${part.comparison_operator}
                    `;

                    // カウンター条件の場合、特別なUIを追加
                    if (condition.kind === 2) { // KindCounter = 2
                        const counterControls = document.createElement('div');
                        counterControls.className = 'counter-controls';
                        // サーバーから取得した現在値を表示するように修正
                        const currentValue = part.current_value !== undefined ? part.current_value : 0;
                        counterControls.innerHTML = `
                            <div class="counter-value">
                                現在値: <span class="current-value">${currentValue}</span> /
                                目標値: <span class="target-value">${part.reference_value_int}</span>
                            </div>
                            <button class="increment-btn" data-condition-id="${condition.id}" data-part-id="${part.id}">
                                カウントアップ
                            </button>
                        `;

                        // カウントアップボタンのイベントリスナーを追加
                        const incrementBtn = counterControls.querySelector('.increment-btn');
                        incrementBtn.addEventListener('click', async () => {
                            try {
                                const result = await this.handleCounterIncrement(condition.id, part.id);
                                const currentValueSpan = counterControls.querySelector('.current-value');
                                currentValueSpan.textContent = result.current_value;
                                
                                if (result.is_satisfied) {
                                    incrementBtn.disabled = true;
                                    this.showStatus('条件を満たしました！', 'success');
                                }
                            } catch (error) {
                                console.error('カウンター更新エラー:', error);
                            }
                        });

                        partBasic.appendChild(counterControls);
                    }
                    
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

        const currentStateElement = document.getElementById('current-state');
        currentStateElement.textContent = data.state;
        currentStateElement.className = `state-display ${data.phase?.is_clear ? 'is-clear' : 'not-clear'}`;
        this.currentState = data.state;

        if (data.phase) {
            document.getElementById('phase-name').textContent = data.phase.name || '-';
            document.getElementById('phase-description').textContent = data.phase.description || '-';
            document.getElementById('phase-order').textContent = data.phase.order || '-';
            document.getElementById('phase-start-time').textContent = this.formatTime(data.phase.start_time);
            document.getElementById('phase-finish-time').textContent = this.formatTime(data.phase.finish_time);
        }

        this.updateStateDiagram(data.state);
        this.updateButtons(data.state);

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
        document.querySelectorAll('.state').forEach(state => {
            state.classList.remove('active');
        });

        document.querySelectorAll('.transition').forEach(transition => {
            transition.classList.remove('active');
        });

        const currentStateElement = document.querySelector(`.state[data-state="${newState}"]`);
        if (currentStateElement) {
            currentStateElement.classList.add('active');
        }

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