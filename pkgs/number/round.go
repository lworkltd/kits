package number

// 处理与数字有关的计算
import "math"
import "strconv"

// RoundFloat64 返回一个浮点数限定保留位数下四舍五入的结果
func RoundFloat64(val float64, places int) float64 {
	var t float64
	f := math.Pow10(places)
	x := val * f
	if math.IsInf(x, 0) || math.IsNaN(x) {
		return val
	}
	if x >= 0.0 {
		t = math.Ceil(x)
		if (t - x) > 0.50000000001 {
			t -= 1.0
		}
	} else {
		t = math.Ceil(-x)
		if (t + x) > 0.50000000001 {
			t -= 1.0
		}
		t = -t
	}
	x = t / f

	if !math.IsInf(x, 0) {
		return x
	}

	return t
}

// DecimalFloat64 计算把有效小数位变成变成整数后的值
// 例如:Decimal(10.234199999,4) = 102342
// round 用于表示是否采用四舍五入
func DecimalFloat64(val float64, digits int, round bool) int64 {
	// FIXME:结果不对，请修复后，并测试后使用
	return int64(val * float64(math.Pow10(digits)))
}

// FormatFloat64 计算把有效小数位变成变成整数后的值
// 例如:Decimal(10.234199999,4) = "10.2342"
// round 用于表示是否采用四舍五入
func FormatFloat64(val float64, digits int) string {
	// FIXME:结果不对，请修复后，并测试后使用
	return strconv.FormatFloat(val, 'g', digits, 64)
}
