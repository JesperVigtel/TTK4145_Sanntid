package types

//Kommentarer til Jesper (meg selv)

//Burde nok flyttes til et annet sted senere...
//Bytt pekere til channels?

type OrderAssignmentService interface {
    AssignOrders(worldview *worldview.Worldview, nodeID int) (OrderSet, error)	
}

type CostCalculator interface {
    CalculateCost(elevator *elevator.Elevator, order *order.Order) int
}

type NetworkService interface {
    Broadcast(message interface{}) error
    Subscribe() <-chan NetworkMessage
}

type HardwareController interface {
    SetMotorDirection(direction elevator.Direction) error
    GetFloorSensor() (int, error)
    SetButtonLight(floor int, button order.ButtonType, on bool) error
}

// application/order_assignment/assigner.go
type Assigner struct {
    costCalculator CostCalculator
    externalAlgo   ExternalAlgorithmClient
}

func NewAssigner(calc CostCalculator, algo ExternalAlgorithmClient) *Assigner {
    return &Assigner{
        costCalculator: calc,
        externalAlgo:   algo,
    }
}