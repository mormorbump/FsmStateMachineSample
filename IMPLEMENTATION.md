# UI改善実装計画

ユーザーからの要望に基づいて、以下のUI改善を実装します：

## 1. リセットボタンを常に押せるようにする

現在の実装では、`script.js`の`updateAutoTransitionStatus`メソッド内で、自動遷移が実行中の場合にリセットボタンが無効化されています。

```javascript
// 現在の実装
updateAutoTransitionStatus(isRunning) {
    console.log('自動遷移状態更新:', isRunning);
    const startBtn = document.getElementById('start-auto');
    const stopBtn = document.getElementById('stop-auto');
    const resetBtn = document.getElementById('reset-btn');
    
    if (isRunning) {
        startBtn.disabled = true;
        stopBtn.disabled = false;
        resetBtn.disabled = true; // リセットボタンが無効化されている
    } else {
        startBtn.disabled = false;
        stopBtn.disabled = true;
        resetBtn.disabled = false;
    }
}
```

この部分を修正して、リセットボタンが常に有効になるようにします：

```javascript
// 修正後の実装
updateAutoTransitionStatus(isRunning) {
    console.log('自動遷移状態更新:', isRunning);
    const startBtn = document.getElementById('start-auto');
    const stopBtn = document.getElementById('stop-auto');
    const resetBtn = document.getElementById('reset-btn');
    
    if (isRunning) {
        startBtn.disabled = true;
        stopBtn.disabled = false;
        // resetBtn.disabled = true; // この行を削除または以下のようにコメントアウト
    } else {
        startBtn.disabled = false;
        stopBtn.disabled = true;
        // resetBtn.disabled = false; // この行も不要になる
    }
    
    // リセットボタンは常に有効
    resetBtn.disabled = false;
}
```

## 2. Phase詳細の横幅を小さくする

現在、`style.css`の`.info-wrapper > div`セレクタで各要素（Phase詳細と条件コンテナ）に等しい幅（flex: 1）が割り当てられています。

```css
/* 現在の実装 */
.info-wrapper > div {
    flex: 1;
}
```

Phase詳細の幅を小さくし、条件コンテナにより多くのスペースを割り当てるように修正します：

```css
/* 修正後の実装 */
.phase-details {
    flex: 0.3; /* 30%の幅に縮小 */
}

.conditions-container {
    flex: 0.7; /* 70%の幅に拡大 */
}
```

## 3. 条件の詳細をスクロールできるようにする

条件コンテナに最大高さを設定し、内容がはみ出した場合にスクロールできるようにします：

```css
/* 追加する実装 */
.conditions-container {
    max-height: 500px; /* 適切な高さに調整 */
    overflow-y: auto; /* 縦方向のスクロールを有効化 */
}

.conditions-list {
    max-height: 450px; /* コンテナより少し小さく */
    overflow-y: auto;
}
```

## 実装手順

1. `script.js`のリセットボタンの動作を修正
2. `style.css`のレイアウトを調整
   - Phase詳細の幅を小さくする
   - 条件コンテナをスクロール可能にする

これらの変更により、ユーザーの要望に応えることができます。