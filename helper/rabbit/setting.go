package rabbit

import (
	"github.com/streadway/amqp"
	"time"
)

const (
	queueSettingPrefix    = "queue/"
	bindSettingPrefix     = "bind/"
	exchangeSettingPrefix = "exchange/"
	consumeSettingPrefix  = "consume/"
)
const (
	settingDurable    = "durable"
	settingExclusive  = "exclusive"
	settingInternal   = "internal"
	settingAutoDelete = "autodelete"
	settingNoWait     = "nowait"
	settingAutoAck    = "autoack"
	settingNoLocal    = "nolocal"
)

func defaultQueueSettings() map[string]bool {
	return map[string]bool{
		settingDurable:    true,
		settingAutoDelete: false,
		settingExclusive:  false,
		settingNoWait:     false,
	}
}

func defaultExchangeSettings() map[string]bool {
	return map[string]bool{
		settingDurable:    true,
		settingAutoDelete: false,
		settingInternal:   false,
		settingNoWait:     false,
	}
}

func defaultConsumeSettings() map[string]bool {
	return map[string]bool{
		settingAutoAck:   true,
		settingExclusive: false,
		settingNoLocal:   false,
		settingNoWait:    false,
	}
}

type QueueSettings map[string]bool

func NewQueueSettings() QueueSettings {
	return map[string]bool{}
}

func (settings QueueSettings) Durable(able bool) QueueSettings {
	settings[queueSettingPrefix+settingDurable] = able
	return settings
}

func (settings QueueSettings) AutoDelete(able bool) QueueSettings {
	settings[queueSettingPrefix+settingAutoDelete] = able

	return settings
}

func (settings QueueSettings) Exclusive(able bool) QueueSettings {
	settings[queueSettingPrefix+settingDurable] = able
	return settings
}

type ExchangeSettings map[string]bool

func NewExchangeSettings() ExchangeSettings {
	return map[string]bool{}
}

func (settings ExchangeSettings) Durable(able bool) ExchangeSettings {
	settings[exchangeSettingPrefix+settingDurable] = able
	return settings
}

func (settings ExchangeSettings) AutoDelete(able bool) ExchangeSettings {
	settings[exchangeSettingPrefix+settingAutoDelete] = able
	return settings
}

func (settings ExchangeSettings) Internal(able bool) ExchangeSettings {
	settings[exchangeSettingPrefix+settingInternal] = able
	return settings
}

type ConsumeSettings map[string]bool

func NewConsumeSettings() ConsumeSettings {
	return map[string]bool{}
}

func (settings ConsumeSettings) AutoAck(able bool) ConsumeSettings {
	settings[consumeSettingPrefix+settingAutoAck] = able
	return settings
}

func (settings ConsumeSettings) Exclusive(able bool) ConsumeSettings {
	settings[consumeSettingPrefix+settingExclusive] = able
	return settings
}

func (settings ConsumeSettings) NoLocal(able bool) ConsumeSettings {
	settings[consumeSettingPrefix+settingNoLocal] = able
	return settings
}

func MakeupSettings(settings ...map[string]bool) map[string]bool {
	allSettings := make(map[string]bool, 10)
	for _, setting := range settings {
		for key, value := range setting {
			allSettings[key] = value
		}
	}
	return allSettings
}

type PublishOption func(*amqp.Publishing)

func OptionContentType(contentType string) PublishOption {
	return func(args *amqp.Publishing) {
		args.ContentType = contentType
	}
}

func OptionContentEncoding(contentEncoding string) PublishOption {
	return func(args *amqp.Publishing) {
		args.ContentEncoding = contentEncoding
	}
}

// OptionDeliveryMode 传输模式
func OptionDeliveryMode(deliveryMode uint8) PublishOption {
	return func(args *amqp.Publishing) {
		args.DeliveryMode = deliveryMode
	}
}

// OptionHeaders 消息头
func OptionHeaders(headers amqp.Table) PublishOption {
	return func(args *amqp.Publishing) {
		args.Headers = headers
	}
}

// OptionPriority 消息优先级
func OptionPriority(priority uint8) PublishOption {
	return func(args *amqp.Publishing) {
		args.Priority = priority
	}
}

// OptionCorrelationId 关联ID
func OptionCorrelationId(correlationId string) PublishOption {
	return func(args *amqp.Publishing) {
		args.CorrelationId = correlationId
	}
}

// OptionReplyTo 响应队列
func OptionReplyTo(replyTo string) PublishOption {
	return func(args *amqp.Publishing) {
		args.ReplyTo = replyTo
	}
}

// OptionExpiration 超时时间
func OptionExpiration(expiration string) PublishOption {
	return func(args *amqp.Publishing) {
		args.Expiration = expiration
	}
}

// OptionMessageId 消息ID
func OptionMessageId(messageId string) PublishOption {
	return func(args *amqp.Publishing) {
		args.MessageId = messageId
	}
}

// OptionTimestamp 不建议使用,配置会使用默认的配置
func OptionTimestamp(timestamp time.Time) PublishOption {
	return func(args *amqp.Publishing) {
		args.Timestamp = timestamp
	}
}

// OptionType 消息的类型
func OptionType(typ string) PublishOption {
	return func(args *amqp.Publishing) {
		args.Type = typ
	}
}

// OptionUserId 用户ID
func OptionUserId(userId string) PublishOption {
	return func(args *amqp.Publishing) {
		args.UserId = userId
	}
}

// OptionAppId 应用ID
func OptionAppId(appId string) PublishOption {
	return func(args *amqp.Publishing) {
		args.AppId = appId
	}
}
