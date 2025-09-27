# E-commerce Order Processing Workflow

This example demonstrates a comprehensive e-commerce order processing system using GoSteps with advanced workflow management, conditional branching, and robust error handling.

## Features Demonstrated

### 1. **Complex Order Workflow**

- Order validation and customer verification
- Inventory management with real-time stock checking
- Payment processing with gateway integration simulation
- Multiple fulfillment strategies based on order characteristics
- Order status tracking and customer communications

### 2. **Advanced Conditional Branching**

Three distinct fulfillment paths based on order characteristics:

**Standard Fulfillment** (Regular orders)
- Standard inventory reservation
- Basic shipping calculation
- Standard shipment scheduling

**Express Fulfillment** (High priority/express shipping)
- Priority inventory reservation
- Express shipping with surcharges
- Expedited shipment scheduling
- Customer notifications for tracking

**Backorder Fulfillment** (Out of stock items)
- Backorder creation and management
- Customer notifications about delays
- Partial fulfillment handling

### 3. **Comprehensive Retry Strategies**

- **Payment Processing**: Retries on gateway timeouts and temporary errors
- **Notifications**: Handles service unavailability with automatic retries
- **Inventory Checks**: Multiple attempts for real-time stock verification
- **Pattern-based Retries**: Uses regex patterns to identify retryable errors

### 4. **Sophisticated Rollback Mechanisms**

- **Inventory Release**: Automatically releases reserved stock on failures
- **Payment Refunds**: Processes refunds when order fails after payment
- **Shipment Cancellation**: Cancels scheduled shipments on downstream failures
- **State Restoration**: Maintains data consistency across rollbacks

### 5. **Real-time Inventory Management**

- Stock level tracking with reservations
- Priority inventory allocation for express orders
- Partial fulfillment handling for mixed availability
- Automatic inventory release on failures

## Workflow Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Validate Order  │ -> │ Check Inventory │ -> │ Process Payment │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                               ┌────────┴────────┐
                                               │ Fulfillment     │
                                               │ Resolver        │
                                               └────────┬────────┘
                        ┌──────────────────────────────┼──────────────────────────────┐
                        │                              │                              │
              ┌─────────▼─────────┐          ┌─────────▼─────────┐          ┌─────────▼─────────┐
              │ Standard          │          │ Express           │          │ Backorder         │
              │ Fulfillment       │          │ Fulfillment       │          │ Fulfillment       │
              │                   │          │                   │          │                   │
              │ • Reserve         │          │ • Priority        │          │ • Create          │
              │   Inventory       │          │   Reserve         │          │   Backorder       │
              │ • Calculate       │          │ • Express         │          │ • Notify          │
              │   Shipping        │          │   Shipping        │          │   Customer        │
              │ • Schedule        │          │ • Express         │          │                   │
              │   Shipment        │          │   Schedule        │          │                   │
              │                   │          │ • Notify          │          │                   │
              │                   │          │   Customer        │          │                   │
              └─────────┬─────────┘          └─────────┬─────────┘          └─────────┬─────────┘
                        │                              │                              │
                        └──────────────────────────────┼──────────────────────────────┘
                                                       │
                                              ┌────────▼────────┐
                                              │ Update Order    │
                                              │ Status          │
                                              └────────┬────────┘
                                                       │
                                              ┌────────▼────────┐
                                              │ Send            │
                                              │ Confirmation    │
                                              └─────────────────┘
```

## Sample Order Data

The example processes orders with the following structure:

```go
Order{
    ID: "ORD-2025-001",
    CustomerID: "CUST-123",
    Items: []OrderItem{
        {ProductID: "PROD-001", Quantity: 2, Price: 29.99, Name: "Wireless Headphones"},
        {ProductID: "PROD-002", Quantity: 1, Price: 49.99, Name: "Bluetooth Speaker"},
    },
    TotalAmount: 109.97,
    PaymentMethod: "credit_card",
    ShippingMethod: "express",
    Priority: "high",
}
```

## Running the Example

```bash
cd example/usecases/ecommerce-order
go run main.go
```

## Key Decision Points

### Fulfillment Routing Logic

1. **Inventory Availability**: If items are out of stock → Backorder Fulfillment
2. **Shipping Method**: If express shipping → Express Fulfillment  
3. **Order Priority**: If high priority → Express Fulfillment
4. **Default**: Standard Fulfillment

### Payment Processing

- Simulates real payment gateway interactions
- Handles temporary gateway timeouts with retries
- Calculates processing fees (2.9%)
- Generates payment references for tracking

### Error Scenarios Handled

- **Validation Failures**: Missing required fields, invalid data
- **Inventory Issues**: Out of stock, insufficient quantities
- **Payment Failures**: Gateway timeouts, processing errors
- **Communication Failures**: Notification service unavailability
- **System Errors**: Database connectivity, external service failures

## Business Value

This example demonstrates how GoSteps can handle:

1. **Complex Business Logic**: Multi-path workflows based on dynamic conditions
2. **Fault Tolerance**: Automatic retries and rollbacks maintain system reliability
3. **Data Consistency**: Transactional behavior across distributed operations
4. **Scalability**: Modular step design allows easy feature additions
5. **Observability**: Comprehensive logging for debugging and monitoring

The system ensures that orders are processed reliably, inventory is managed accurately, and customers receive appropriate communications throughout the fulfillment process.