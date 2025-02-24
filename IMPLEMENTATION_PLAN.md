# isClearフィールド追加の実装計画

## 概要
Phase、Condition、ConditionPartにisClearというカラムを追加し、条件が満たされた時に適切に更新する実装を行います。

## 変更対象
1. ConditionPart
   - isClearフィールドの追加
   - StateSatisfiedになった時にisClearをtrueに設定
   - テストケースの追加

2. Condition
   - isClearフィールドの追加
   - OnConditionSatisfiedでsatisfiedがtrueの時にisClearをtrueに設定
   - テストケースの追加

3. Phase
   - isClearフィールドの追加
   - OnConditionSatisfiedでsatisfiedがtrueの時にisClearをtrueに設定
   - テストケースの追加

## 実装ステップ

### 1. ConditionPartの実装
1. internal/domain/entity/condition_part.go の修正
   - structにisClearフィールドを追加
   - NewConditionPartでisClearをfalseで初期化
   - IsClearメソッドの追加
   - StateSatisfiedになった時にisClearをtrueに設定するロジックの追加

2. internal/domain/entity/condition_part_test.go の修正
   - isClearの初期値テスト
   - StateSatisfiedになった時のisClearの変更テスト
   - IsClearメソッドのテスト

### 2. Conditionの実装
1. internal/domain/entity/condition.go の修正
   - structにisClearフィールドを追加
   - NewConditionでisClearをfalseで初期化
   - IsClearメソッドの追加
   - OnConditionSatisfiedでsatisfiedがtrueの時にisClearをtrueに設定するロジックの追加

2. internal/domain/entity/condition_test.go の修正
   - isClearの初期値テスト
   - OnConditionSatisfiedでsatisfiedがtrueの時のisClearの変更テスト
   - IsClearメソッドのテスト

### 3. Phaseの実装
1. internal/domain/entity/phase.go の修正
   - structにisClearフィールドを追加
   - NewPhaseでisClearをfalseで初期化
   - IsClearメソッドの追加
   - OnConditionSatisfiedでsatisfiedがtrueの時にisClearをtrueに設定するロジックの追加

2. internal/domain/entity/phase_test.go の修正
   - isClearの初期値テスト
   - OnConditionSatisfiedでsatisfiedがtrueの時のisClearの変更テスト
   - IsClearメソッドのテスト

## テスト計画

### ConditionPartのテスト
1. 初期状態でisClearがfalseであることを確認
2. StateSatisfiedになった時にisClearがtrueになることを確認
3. IsClearメソッドが正しい値を返すことを確認

### Conditionのテスト
1. 初期状態でisClearがfalseであることを確認
2. OnConditionSatisfiedでsatisfiedがtrueの時にisClearがtrueになることを確認
3. IsClearメソッドが正しい値を返すことを確認

### Phaseのテスト
1. 初期状態でisClearがfalseであることを確認
2. OnConditionSatisfiedでsatisfiedがtrueの時にisClearがtrueになることを確認
3. IsClearメソッドが正しい値を返すことを確認

## 実装の注意点
- 各エンティティのisClearは一度trueになったら、falseに戻らない
- テストは各状態遷移を確実にカバーする
- 既存の機能に影響を与えないように注意する

## 実装順序
1. ConditionPart
2. Condition
3. Phase

この順序で実装することで、依存関係の低い方から順に実装を進めることができます。