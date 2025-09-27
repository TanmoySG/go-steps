package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	gosteps "github.com/TanmoySG/go-steps"
	"github.com/rs/zerolog"
)

// Order represents an e-commerce order
type Order struct {
	ID                string         `json:"id"`
	CustomerID        string         `json:"customer_id"`
	Items             []OrderItem    `json:"items"`
	TotalAmount       float64        `json:"total_amount"`
	PaymentMethod     string         `json:"payment_method"`
	ShippingAddress   Address        `json:"shipping_address"`
	Status            string         `json:"status"`
	CreatedAt         time.Time      `json:"created_at"`
	Priority          string         `json:"priority"`
	ShippingMethod    string         `json:"shipping_method"`
	InventoryReserved map[string]int `json:"inventory_reserved"`
	PaymentReference  string         `json:"payment_reference"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Name      string  `json:"name"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZIP     string `json:"zip"`
	Country string `json:"country"`
}

type InventoryItem struct {
	ProductID string `json:"product_id"`
	Stock     int    `json:"stock"`
	Reserved  int    `json:"reserved"`
}

func main() {
	// Setup logging
	runLogFile, _ := os.OpenFile(
		"order-processing.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	output := zerolog.MultiLevelWriter(os.Stdout, runLogFile)
	logger := gosteps.NewGoStepsLogger(output, &gosteps.LoggerOpts{StepLoggingEnabled: true})

	// Initialize context with sample order
	ctx := gosteps.NewGoStepsContext()
	ctx.Use(logger)

	sampleOrder := Order{
		ID:         "ORD-2025-001",
		CustomerID: "CUST-123",
		Items: []OrderItem{
			{ProductID: "PROD-001", Quantity: 2, Price: 29.99, Name: "Wireless Headphones"},
			{ProductID: "PROD-002", Quantity: 1, Price: 49.99, Name: "Bluetooth Speaker"},
		},
		TotalAmount:       109.97,
		PaymentMethod:     "credit_card",
		ShippingMethod:    "express",
		Priority:          "high",
		Status:            "pending",
		CreatedAt:         time.Now(),
		InventoryReserved: make(map[string]int),
		ShippingAddress: Address{
			Street:  "123 Main St",
			City:    "Seattle",
			State:   "WA",
			ZIP:     "98101",
			Country: "US",
		},
	}

	// Initialize inventory
	inventory := map[string]InventoryItem{
		"PROD-001": {ProductID: "PROD-001", Stock: 10, Reserved: 0},
		"PROD-002": {ProductID: "PROD-002", Stock: 5, Reserved: 0},
		"PROD-003": {ProductID: "PROD-003", Stock: 0, Reserved: 0}, // Out of stock
	}

	ctx.WithData(map[string]interface{}{
		"order":        sampleOrder,
		"inventory":    inventory,
		"paymentFees":  0.0,
		"shippingCost": 0.0,
	})

	// Define the order processing workflow
	steps := gosteps.Steps{
		{
			Name:     "validateOrder",
			Function: validateOrderStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
				RetrySleep:     1 * time.Second,
			},
		},
		{
			Name:     "checkInventory",
			Function: checkInventoryStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				RetrySleep:     2 * time.Second,
			},
			RollbackFunction: releaseInventoryReservation,
		},
		{
			Name:     "processPayment",
			Function: processPaymentStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				ErrorPatternsToRetry: []regexp.Regexp{
					*regexp.MustCompile("payment.*gateway.*timeout"),
					*regexp.MustCompile("temporary.*payment.*error"),
				},
				RetrySleep: 3 * time.Second,
			},
			RollbackFunction: refundPayment,
			Branches: &gosteps.Branches{
				Resolver: fulfillmentResolver,
				Branches: []gosteps.Branch{
					{
						BranchName: "standardFulfillment",
						Steps: gosteps.Steps{
							{
								Name:             "reserveInventory",
								Function:         reserveInventoryStep,
								RollbackFunction: releaseInventoryReservation,
							},
							{
								Name:     "calculateShipping",
								Function: calculateShippingStep,
							},
							{
								Name:             "scheduleShipment",
								Function:         scheduleShipmentStep,
								RollbackFunction: cancelShipment,
							},
						},
					},
					{
						BranchName: "expressFulfillment",
						Steps: gosteps.Steps{
							{
								Name:             "reservePriorityInventory",
								Function:         reservePriorityInventoryStep,
								RollbackFunction: releaseInventoryReservation,
							},
							{
								Name:     "calculateExpressShipping",
								Function: calculateExpressShippingStep,
							},
							{
								Name:             "scheduleExpressShipment",
								Function:         scheduleExpressShipmentStep,
								RollbackFunction: cancelShipment,
							},
							{
								Name:     "notifyCustomer",
								Function: notifyCustomerStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     1 * time.Second,
								},
							},
						},
					},
					{
						BranchName: "backorderFulfillment",
						Steps: gosteps.Steps{
							{
								Name:     "createBackorder",
								Function: createBackorderStep,
							},
							{
								Name:     "notifyBackorder",
								Function: notifyBackorderStep,
								StepOpts: gosteps.StepOpts{
									MaxRunAttempts: 2,
									RetrySleep:     1 * time.Second,
								},
							},
						},
					},
				},
			},
		},
		{
			Name:     "updateOrderStatus",
			Function: updateOrderStatusStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 2,
			},
		},
		{
			Name:     "sendConfirmation",
			Function: sendConfirmationStep,
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 3,
				RetrySleep:     2 * time.Second,
			},
		},
	}

	// Execute the order processing workflow
	fmt.Println("🛒 Starting E-commerce Order Processing...")
	stepsProcessor := gosteps.NewStepsProcessor(steps)
	stepsProcessor.Execute(ctx)

	// Print final order status
	finalOrder := ctx.GetData("order").(Order)
	fmt.Printf("\n📦 Order Processing Complete!\n")
	fmt.Printf("Order ID: %s\n", finalOrder.ID)
	fmt.Printf("Status: %s\n", finalOrder.Status)
	fmt.Printf("Total Amount: $%.2f\n", finalOrder.TotalAmount)
	if finalOrder.PaymentReference != "" {
		fmt.Printf("Payment Reference: %s\n", finalOrder.PaymentReference)
	}
}

// Step Functions

func validateOrderStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Validating order details...")

	order := ctx.GetData("order").(Order)

	// Validate order fields
	if order.ID == "" {
		return gosteps.MarkStateFailed().WithMessage("Order ID is required")
	}

	if order.CustomerID == "" {
		return gosteps.MarkStateFailed().WithMessage("Customer ID is required")
	}

	if len(order.Items) == 0 {
		return gosteps.MarkStateFailed().WithMessage("Order must contain at least one item")
	}

	if order.TotalAmount <= 0 {
		return gosteps.MarkStateFailed().WithMessage("Order total must be greater than zero")
	}

	// Validate shipping address
	if order.ShippingAddress.Street == "" || order.ShippingAddress.City == "" {
		return gosteps.MarkStateFailed().WithMessage("Complete shipping address is required")
	}

	// Set order status to validated
	order.Status = "validated"
	ctx.SetData("order", order)

	return gosteps.MarkStateComplete().WithMessage("order validation successful")
}

func checkInventoryStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Checking inventory availability...")

	order := ctx.GetData("order").(Order)
	inventory := ctx.GetData("inventory").(map[string]InventoryItem)

	var unavailableItems []string
	allAvailable := true

	for _, item := range order.Items {
		inventoryItem, exists := inventory[item.ProductID]
		if !exists {
			unavailableItems = append(unavailableItems, item.ProductID)
			allAvailable = false
			continue
		}

		availableStock := inventoryItem.Stock - inventoryItem.Reserved
		if availableStock < item.Quantity {
			unavailableItems = append(unavailableItems,
				fmt.Sprintf("%s (need %d, have %d)", item.ProductID, item.Quantity, availableStock))
			allAvailable = false
		}
	}

	if !allAvailable {
		// Store unavailable items for backorder processing
		ctx.SetData("unavailableItems", unavailableItems)

		// Check if any items are available for partial fulfillment
		hasPartialInventory := false
		for _, item := range order.Items {
			if inventoryItem, exists := inventory[item.ProductID]; exists {
				if inventoryItem.Stock > inventoryItem.Reserved {
					hasPartialInventory = true
					break
				}
			}
		}

		if hasPartialInventory {
			return gosteps.MarkStateComplete().WithMessage("Partial inventory available - will process backorder")
		} else {
			return gosteps.MarkStateComplete().WithMessage("No inventory available - will create full backorder")
		}
	}

	return gosteps.MarkStateComplete().WithMessage("All items in stock and available")
}

func processPaymentStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Processing payment...")

	order := ctx.GetData("order").(Order)

	// Simulate payment processing with random failures
	rand.Seed(time.Now().UnixNano())

	// Simulate payment gateway timeout (retry)
	if rand.Float32() < 0.2 {
		return gosteps.MarkStateError().WithError(fmt.Errorf("payment gateway timeout - temporary error"))
	}

	// Simulate payment method validation
	if order.PaymentMethod == "" {
		return gosteps.MarkStateFailed().WithMessage("Payment method is required")
	}

	// Calculate payment fees
	paymentFees := order.TotalAmount * 0.029 // 2.9% processing fee
	ctx.SetData("paymentFees", paymentFees)

	// Simulate payment processing success
	paymentReference := fmt.Sprintf("PAY-%d", time.Now().UnixNano())
	order.PaymentReference = paymentReference
	order.Status = "payment_processed"

	ctx.SetData("order", order)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Payment processed successfully. Reference: %s", paymentReference)).
		WithData(map[string]interface{}{
			"paymentReference": paymentReference,
			"paymentFees":      paymentFees,
		})
}

func fulfillmentResolver(ctx gosteps.GoStepsCtx) gosteps.BranchName {
	order := ctx.GetData("order").(Order)
	unavailableItems := ctx.GetData("unavailableItems")

	// Check if we have unavailable items (backorder scenario)
	if unavailableItems != nil {
		items := unavailableItems.([]string)
		if len(items) > 0 {
			ctx.Log("Some items unavailable - routing to backorder fulfillment")
			return gosteps.BranchName("backorderFulfillment")
		}
	}

	// Route based on shipping method and priority
	if order.ShippingMethod == "express" || order.Priority == "high" {
		ctx.Log("High priority order - routing to express fulfillment")
		return gosteps.BranchName("expressFulfillment")
	}

	ctx.Log("Standard order - routing to standard fulfillment")
	return gosteps.BranchName("standardFulfillment")
}

func reserveInventoryStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Reserving inventory for standard fulfillment...")

	order := ctx.GetData("order").(Order)
	inventory := ctx.GetData("inventory").(map[string]InventoryItem)

	for _, item := range order.Items {
		inventoryItem := inventory[item.ProductID]
		inventoryItem.Reserved += item.Quantity
		inventory[item.ProductID] = inventoryItem

		order.InventoryReserved[item.ProductID] = item.Quantity
	}

	ctx.SetData("inventory", inventory)
	ctx.SetData("order", order)

	return gosteps.MarkStateComplete().WithMessage("Inventory reserved successfully")
}

func reservePriorityInventoryStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Reserving priority inventory for express fulfillment...")

	order := ctx.GetData("order").(Order)
	inventory := ctx.GetData("inventory").(map[string]InventoryItem)

	for _, item := range order.Items {
		inventoryItem := inventory[item.ProductID]
		inventoryItem.Reserved += item.Quantity
		inventory[item.ProductID] = inventoryItem

		order.InventoryReserved[item.ProductID] = item.Quantity
	}

	ctx.SetData("inventory", inventory)
	ctx.SetData("order", order)

	return gosteps.MarkStateComplete().WithMessage("Priority inventory reserved successfully")
}

func calculateShippingStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Calculating standard shipping cost...")

	order := ctx.GetData("order").(Order)

	// Calculate shipping based on weight and distance
	baseShipping := 9.99
	weightSurcharge := float64(len(order.Items)) * 2.50
	shippingCost := baseShipping + weightSurcharge

	ctx.SetData("shippingCost", shippingCost)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Standard shipping calculated: $%.2f", shippingCost)).
		WithData(map[string]interface{}{
			"shippingCost": shippingCost,
		})
}

func calculateExpressShippingStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Calculating express shipping cost...")

	order := ctx.GetData("order").(Order)

	// Express shipping is more expensive
	baseShipping := 24.99
	weightSurcharge := float64(len(order.Items)) * 3.99
	expressSurcharge := 10.00
	shippingCost := baseShipping + weightSurcharge + expressSurcharge

	ctx.SetData("shippingCost", shippingCost)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Express shipping calculated: $%.2f", shippingCost)).
		WithData(map[string]interface{}{
			"shippingCost": shippingCost,
		})
}

func scheduleShipmentStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Scheduling standard shipment...")

	order := ctx.GetData("order").(Order)

	// Schedule shipment for 2-3 business days
	estimatedDelivery := time.Now().AddDate(0, 0, 3)
	order.Status = "scheduled_for_shipment"

	ctx.SetData("order", order)
	ctx.SetData("estimatedDelivery", estimatedDelivery)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Standard shipment scheduled for %s", estimatedDelivery.Format("2006-01-02")))
}

func scheduleExpressShipmentStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Scheduling express shipment...")

	order := ctx.GetData("order").(Order)

	// Express delivery within 1-2 business days
	estimatedDelivery := time.Now().AddDate(0, 0, 1)
	order.Status = "scheduled_for_express_shipment"

	ctx.SetData("order", order)
	ctx.SetData("estimatedDelivery", estimatedDelivery)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Express shipment scheduled for %s", estimatedDelivery.Format("2006-01-02")))
}

func createBackorderStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Creating backorder for unavailable items...")

	order := ctx.GetData("order").(Order)
	unavailableItems := ctx.GetData("unavailableItems").([]string)

	backorderID := fmt.Sprintf("BO-%d", time.Now().UnixNano())
	order.Status = "backordered"

	ctx.SetData("order", order)
	ctx.SetData("backorderID", backorderID)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Backorder created: %s for items: %v", backorderID, unavailableItems)).
		WithData(map[string]interface{}{
			"backorderID": backorderID,
		})
}

func notifyCustomerStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Sending customer notification...")

	order := ctx.GetData("order").(Order)
	estimatedDelivery := ctx.GetData("estimatedDelivery").(time.Time)

	// Simulate notification sending with potential failure
	if rand.Float32() < 0.1 {
		return gosteps.MarkStatePending().WithMessage("Notification service temporarily unavailable")
	}

	notificationID := fmt.Sprintf("NOTIF-%d", time.Now().UnixNano())

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Customer %s notified for order %s. Delivery: %s. ID: %s",
			order.CustomerID, order.ID, estimatedDelivery.Format("2006-01-02"), notificationID)).
		WithData(map[string]interface{}{
			"notificationID": notificationID,
		})
}

func notifyBackorderStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Sending backorder notification...")

	order := ctx.GetData("order").(Order)
	backorderID := ctx.GetData("backorderID").(string)

	// Simulate backorder notification
	notificationID := fmt.Sprintf("BACKORDER-NOTIF-%d", time.Now().UnixNano())

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Backorder notification sent to customer %s for order %s (backorder %s). ID: %s",
			order.CustomerID, order.ID, backorderID, notificationID)).
		WithData(map[string]interface{}{
			"backorderNotificationID": notificationID,
		})
}

func updateOrderStatusStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Updating final order status...")

	order := ctx.GetData("order").(Order)

	// Update to final status based on current status
	switch order.Status {
	case "scheduled_for_shipment":
		order.Status = "processing"
	case "scheduled_for_express_shipment":
		order.Status = "express_processing"
	case "backordered":
		order.Status = "awaiting_inventory"
	default:
		order.Status = "confirmed"
	}

	ctx.SetData("order", order)

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Order status updated to: %s", order.Status))
}

func sendConfirmationStep(ctx gosteps.GoStepsCtx) gosteps.StepResult {
	ctx.Log("Sending order confirmation...")

	order := ctx.GetData("order").(Order)

	// Simulate confirmation email sending
	confirmationID := fmt.Sprintf("CONF-%d", time.Now().UnixNano())

	return gosteps.MarkStateComplete().
		WithMessage(fmt.Sprintf("Order confirmation sent to customer %s for order %s ($%.2f). ID: %s",
			order.CustomerID, order.ID, order.TotalAmount, confirmationID)).
		WithData(map[string]interface{}{
			"confirmationID": confirmationID,
		})
}

// Rollback Functions

func releaseInventoryReservation(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Releasing inventory reservations...")

	order := ctx.GetData("order").(Order)
	inventory := ctx.GetData("inventory").(map[string]InventoryItem)

	// Release reserved inventory
	for productID, reservedQty := range order.InventoryReserved {
		if inventoryItem, exists := inventory[productID]; exists {
			inventoryItem.Reserved -= reservedQty
			if inventoryItem.Reserved < 0 {
				inventoryItem.Reserved = 0
			}
			inventory[productID] = inventoryItem
		}
	}

	// Clear reservations from order
	order.InventoryReserved = make(map[string]int)
	ctx.SetData("order", order)
	ctx.SetData("inventory", inventory)

	message := "Inventory reservations released successfully"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func refundPayment(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Processing payment refund...")

	order := ctx.GetData("order").(Order)

	if order.PaymentReference == "" {
		message := "No payment to refund"
		return gosteps.RollbackResult{
			RollbackState:   gosteps.RollbackStateSuccess,
			RollbackMessage: &message,
		}
	}

	// Simulate refund processing
	refundID := fmt.Sprintf("REFUND-%d", time.Now().UnixNano())

	// Clear payment reference
	order.PaymentReference = ""
	order.Status = "refunded"
	ctx.SetData("order", order)

	message := fmt.Sprintf("Payment refunded successfully. Refund ID: %s", refundID)
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}

func cancelShipment(ctx gosteps.GoStepsCtx) gosteps.RollbackResult {
	ctx.Log("Canceling shipment...")

	order := ctx.GetData("order").(Order)

	// Update order status
	order.Status = "shipment_cancelled"
	ctx.SetData("order", order)

	// Remove estimated delivery
	ctx.SetData("estimatedDelivery", nil)

	message := "Shipment cancelled successfully"
	return gosteps.RollbackResult{
		RollbackState:   gosteps.RollbackStateSuccess,
		RollbackMessage: &message,
	}
}
