package dispatch

import (
	. "elevator/internal/types"
)

// Elevator formatting has been moved to internal/types for better cohesion (CC §5.2).
// The conversion logic now lives alongside the type definitions.
func toHRAElevState(elev Elevator) HRAElevState {
	return NewHRAElevState(elev)
}
