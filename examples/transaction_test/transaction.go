package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/libsv/go-bk/wif"
	"github.com/sCrypt-Inc/go-bt/v2"
	"github.com/sCrypt-Inc/go-bt/v2/bscript"
	"github.com/sCrypt-Inc/go-bt/v2/unlocker"
)

func main() {
	fmt.Println("=== TBC Transaction 测试程序 ===\n")

	// 示例1: 基本交易创建
	basicTransaction()

	// 示例2: 使用 UTXO 创建交易
	transactionFromUTXOs()

	// 示例3: 带找零地址的交易
	transactionWithChange()

	// 示例4: 获取输入和输出总额
	transactionAmounts()

	// 示例5: 交易序列化
	transactionSerialization()

	// 示例6: 手续费相关
	feeExamples()

	// 示例7: 时间锁定交易
	timeLockedTransaction()

	// 示例8: OP_RETURN 输出
	transactionWithOpReturn()

	fmt.Println("\n=== 所有测试完成 ===")
}

// 示例1: 基本交易创建
// 对应文档中的: new Transaction().from(utxos).to(address, amount).change(address).sign(privkeySet)
func basicTransaction() {
	fmt.Println("--- 示例1: 基本交易创建 ---")

	tx := bt.NewTx()

	// 添加输入 - 从之前的交易输出
	err := tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d", // 之前的交易ID
		0,                                                                    // 输出索引
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac6a0568656c6c6f",     // 锁定脚本（十六进制）
		1500,                                                                 // satoshis 数量
	)
	if err != nil {
		log.Printf("添加输入失败: %v", err)
		return
	}

	// 添加输出 - 支付到地址
	err = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)
	if err != nil {
		log.Printf("添加输出失败: %v", err)
		return
	}

	// 设置找零地址并计算手续费
	feeQuote := bt.NewFeeQuote()
	err = tx.ChangeToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", feeQuote)
	if err != nil {
		log.Printf("设置找零地址失败: %v", err)
	}

	// 签名交易
	decodedWif, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
	if err != nil {
		log.Printf("解码 WIF 失败: %v", err)
		return
	}

	err = tx.FillAllInputs(context.Background(), &unlocker.Getter{PrivateKey: decodedWif.PrivKey})
	if err != nil {
		log.Printf("签名失败: %v", err)
		return
	}

	fmt.Printf("交易ID: %s\n", tx.TxID())
	fmt.Printf("序列化交易: %s\n\n", tx.String())
}

// 示例2: 使用 UTXO 创建交易
// 对应文档中的: from(utxos) - 传入 UTXO 数组
func transactionFromUTXOs() {
	fmt.Println("--- 示例2: 使用 UTXO 创建交易 ---")

	tx := bt.NewTx()

	// 创建 UTXO
	utxo1 := &bt.UTXO{
		TxID:          mustDecodeHex("11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d"),
		Vout:          0,
		Satoshis:      1000,
		LockingScript: mustLockingScript("76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac"),
	}

	utxo2 := &bt.UTXO{
		TxID:          mustDecodeHex("b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576"),
		Vout:          0,
		Satoshis:      2000,
		LockingScript: mustLockingScript("76a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac"),
	}

	// 从多个 UTXO 添加输入
	err := tx.FromUTXOs(utxo1, utxo2)
	if err != nil {
		log.Printf("从 UTXO 添加输入失败: %v", err)
		return
	}

	// 添加输出
	err = tx.PayToAddress("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 2500)
	if err != nil {
		log.Printf("添加输出失败: %v", err)
		return
	}

	fmt.Printf("输入数量: %d\n", tx.InputCount())
	fmt.Printf("输出数量: %d\n", tx.OutputCount())
	fmt.Printf("输入总额: %d satoshis\n", tx.TotalInputSatoshis())
	fmt.Printf("输出总额: %d satoshis\n\n", tx.TotalOutputSatoshis())
}

// 示例3: 带找零地址的交易
// 对应文档中的: change(address) - 设置找零地址
func transactionWithChange() {
	fmt.Println("--- 示例3: 带找零地址的交易 ---")

	tx := bt.NewTx()

	// 添加输入 (1500 satoshis)
	err := tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		1500,
	)
	if err != nil {
		log.Printf("添加输入失败: %v", err)
		return
	}

	// 添加输出 (1000 satoshis)
	err = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)
	if err != nil {
		log.Printf("添加输出失败: %v", err)
		return
	}

	// 设置找零地址 - 会自动计算手续费并将剩余金额作为找零
	feeQuote := bt.NewFeeQuote()
	err = tx.ChangeToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", feeQuote)
	if err != nil {
		log.Printf("设置找零地址失败: %v", err)
	} else {
		fmt.Printf("找零已自动添加到输出中\n")
	}

	fmt.Printf("输入总额: %d satoshis\n", tx.TotalInputSatoshis())
	fmt.Printf("输出总额: %d satoshis\n", tx.TotalOutputSatoshis())
	fmt.Printf("输出数量: %d\n\n", tx.OutputCount())
}

// 示例4: 获取输入和输出总额
// 对应文档中的: inputAmount 和 outputAmount 字段
func transactionAmounts() {
	fmt.Println("--- 示例4: 获取输入和输出总额 ---")

	tx := bt.NewTx()

	// 添加多个输入
	err := tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		1500,
	)
	if err == nil {
		err = tx.From(
			"b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576",
			0,
			"76a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac",
			2000,
		)
	}

	if err != nil {
		log.Printf("添加输入失败: %v", err)
		return
	}

	// 添加多个输出
	_ = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)
	_ = tx.PayToAddress("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 1500)

	inputAmount := tx.TotalInputSatoshis()
	outputAmount := tx.TotalOutputSatoshis()

	fmt.Printf("输入总额 (inputAmount): %d satoshis\n", inputAmount)
	fmt.Printf("输出总额 (outputAmount): %d satoshis\n", outputAmount)
	fmt.Printf("差额: %d satoshis\n\n", inputAmount-outputAmount)
}

// 示例5: 交易序列化
// 对应文档中的: serialize(), toObject(), toJSON(), toString(), toBuffer()
func transactionSerialization() {
	fmt.Println("--- 示例5: 交易序列化 ---")

	tx := bt.NewTx()

	_ = tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		1500,
	)

	_ = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)

	// toString() - 返回十六进制序列化字符串
	serializedHex := tx.String()
	fmt.Printf("序列化 (toString): %s\n", serializedHex)

	// Bytes() - 返回字节数组
	serializedBytes := tx.Bytes()
	fmt.Printf("序列化 (Bytes): %d 字节\n", len(serializedBytes))

	// NodeJSON() - 返回 JSON 格式的对象表示
	nodeJSON := tx.NodeJSON()
	fmt.Printf("JSON 格式可用: %v\n", nodeJSON != nil)

	// 交易ID
	fmt.Printf("交易ID: %s\n\n", tx.TxID())
}

// 示例6: 手续费相关
// 对应文档中的: fee(), getFee(), ChangeToAddress() 使用 FeeQuote
func feeExamples() {
	fmt.Println("--- 示例6: 手续费相关 ---")

	tx := bt.NewTx()

	_ = tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		1500,
	)

	_ = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)

	// 创建手续费报价
	feeQuote := bt.NewFeeQuote()

	// 使用 ChangeToAddress 会自动计算手续费并将剩余作为找零
	err := tx.ChangeToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", feeQuote)
	if err != nil {
		log.Printf("设置找零失败: %v", err)
		return
	}

	// 估算交易大小
	size, err := tx.EstimateSize()
	if err != nil {
		log.Printf("估算大小失败: %v", err)
	} else {
		fmt.Printf("估算交易大小: %d 字节\n", size)
	}

	// 检查手续费是否足够
	isEnough, err := tx.EstimateIsFeePaidEnough(feeQuote)
	if err != nil {
		log.Printf("检查手续费失败: %v", err)
	} else {
		fmt.Printf("手续费是否足够: %v\n", isEnough)
	}

	// 估算手续费
	fees, err := tx.EstimateFeesPaid(feeQuote)
	if err != nil {
		log.Printf("估算手续费失败: %v", err)
	} else {
		fmt.Printf("估算手续费: %d satoshis\n", fees.TotalFeePaid)
	}

	fmt.Printf("输入总额: %d satoshis\n", tx.TotalInputSatoshis())
	fmt.Printf("输出总额: %d satoshis\n\n", tx.TotalOutputSatoshis())
}

// 示例7: 时间锁定交易
// 对应文档中的: lockUntilDate(), lockUntilBlockHeight(), getLockTime()
func timeLockedTransaction() {
	fmt.Println("--- 示例7: 时间锁定交易 ---")

	tx := bt.NewTx()

	// 设置锁定时间 - 使用未来的日期
	futureDate := time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC)
	
	// 将时间戳转换为 locktime (Unix 时间戳)
	// LockTime 字段如果是时间戳，必须 >= 500000000
	locktime := uint32(futureDate.Unix())
	if locktime < 500000000 {
		locktime = uint32(futureDate.Unix()) + 500000000
	}
	tx.LockTime = locktime

	_ = tx.From(
		"11b476ad8e0a48fcd40807a111a050af51114877e09283bfa7f3505081a1819d",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		1500,
	)

	_ = tx.PayToAddress("1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb", 1000)

	fmt.Printf("锁定时间 (LockTime): %d\n", tx.LockTime)
	fmt.Printf("对应日期: %s\n", futureDate.Format(time.RFC3339))
	
	// 检查 LockTime 是否表示时间戳 (> 500000000) 还是区块高度
	if tx.LockTime >= 500000000 {
		lockTimeDate := time.Unix(int64(tx.LockTime), 0)
		fmt.Printf("解析后的锁定日期: %s\n", lockTimeDate.Format(time.RFC3339))
	} else {
		fmt.Printf("锁定区块高度: %d\n", tx.LockTime)
	}

	fmt.Println()
}

// 示例8: OP_RETURN 输出
// 对应文档中的: 添加数据输出
func transactionWithOpReturn() {
	fmt.Println("--- 示例8: OP_RETURN 输出 ---")

	tx := bt.NewTx()

	_ = tx.From(
		"b7b0650a7c3a1bd4716369783876348b59f5404784970192cec1996e86950576",
		0,
		"76a9149cbe9f5e72fa286ac8a38052d1d5337aa363ea7f88ac",
		1000,
	)

	_ = tx.PayToAddress("1C8bzHM8XFBHZ2ZZVvFy2NSoAZbwCXAicL", 900)

	// 添加 OP_RETURN 输出（数据输出）
	data := []byte("You are using go-bt!")
	err := tx.AddOpReturnOutput(data)
	if err != nil {
		log.Printf("添加 OP_RETURN 输出失败: %v", err)
		return
	}

	fmt.Printf("输出数量: %d\n", tx.OutputCount())
	fmt.Printf("包含数据输出: %v\n", tx.HasDataOutputs())
	fmt.Printf("交易序列化: %s\n\n", tx.String())
}

// 辅助函数：解码十六进制字符串
func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// 辅助函数：从十六进制字符串创建锁定脚本
func mustLockingScript(hexStr string) *bscript.Script {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return bscript.NewFromBytes(b)
}

