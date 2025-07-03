package banktransfer

const (
	TaskQueueName              = "POST_PAYMENT_ORDER_TASK_QUEUE"
	WorkflowName               = "ProcessPostPaymentOrder"
	WorkflowPrePaymentPrefix   = "order_pre_payment_"
	WorkflowPostPaymentPrefix  = "order_post_payment_"
	SignalNamePaymentCompleted = "payment-completed"
)
