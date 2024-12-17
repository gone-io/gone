package gone

import (
	"testing"
)

// createTestCoffin is a helper function to create test coffins
func createTestCoffin(name string) *coffin {
	return &coffin{
		name:  name,
		goner: struct{}{},
	}
}

func TestCheckCircularDepsAndGetBestInitOrder(t *testing.T) {
	tests := []struct {
		name             string
		initiatorDepsMap map[*coffin][]*coffin
		wantCircular     bool
		wantOrderLen     int
	}{
		{
			name: "Linear dependency chain",
			initiatorDepsMap: func() map[*coffin][]*coffin {
				a := createTestCoffin("A")
				b := createTestCoffin("B")
				c := createTestCoffin("C")
				return map[*coffin][]*coffin{
					a: {b},
					b: {c},
					c: {},
				}
			}(),
			wantCircular: false,
			wantOrderLen: 3,
		},
		{
			name: "Circular dependency",
			initiatorDepsMap: func() map[*coffin][]*coffin {
				a := createTestCoffin("A")
				b := createTestCoffin("B")
				c := createTestCoffin("C")
				return map[*coffin][]*coffin{
					a: {b},
					b: {c},
					c: {a}, // Creates a cycle
				}
			}(),
			wantCircular: true,
			wantOrderLen: 0,
		},
		{
			name:             "Empty dependency map",
			initiatorDepsMap: map[*coffin][]*coffin{},
			wantCircular:     false,
			wantOrderLen:     0,
		},
		{
			name: "Diamond dependency",
			initiatorDepsMap: func() map[*coffin][]*coffin {
				a := createTestCoffin("A")
				b := createTestCoffin("B")
				c := createTestCoffin("C")
				d := createTestCoffin("D")
				return map[*coffin][]*coffin{
					a: {b, c},
					b: {d},
					c: {d},
					d: {},
				}
			}(),
			wantCircular: false,
			wantOrderLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circularDeps, initOrder := checkCircularDepsAndGetBestInitOrder(tt.initiatorDepsMap)

			// Check circular dependency detection
			if (len(circularDeps) > 0) != tt.wantCircular {
				t.Errorf("checkCircularDepsAndGetBestInitOrder() circular = %v, want %v",
					len(circularDeps) > 0, tt.wantCircular)
			}

			// Check initialization order length
			if len(initOrder) != tt.wantOrderLen {
				t.Errorf("checkCircularDepsAndGetBestInitOrder() order length = %v, want %v",
					len(initOrder), tt.wantOrderLen)
			}

			if len(initOrder) > 0 {
				// Verify the initialization order is valid
				seen := make(map[*coffin]bool)
				for _, co := range initOrder {
					// Check that all dependencies of current coffin have been initialized
					for _, dep := range tt.initiatorDepsMap[co] {
						if !seen[dep] {
							t.Errorf("Invalid initialization order: %v depends on %v but it's not initialized yet",
								co.name, dep.name)
						}
					}
					seen[co] = true
				}
			}
		})
	}
}
