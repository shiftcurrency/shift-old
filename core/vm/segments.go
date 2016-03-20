package vm

import "math/big"

type jumpSeg struct {
	pos uint64
	err error
	nrg *big.Int
}

func (j jumpSeg) do(program *Program, pc *uint64, env Environment, contract *Contract, memory *Memory, stack *stack) ([]byte, error) {
	if !contract.UseNrg(j.nrg) {
		return nil, OutOfNrgError
	}
	if j.err != nil {
		return nil, j.err
	}
	*pc = j.pos
	return nil, nil
}
func (s jumpSeg) halts() bool { return false }
func (s jumpSeg) Op() OpCode  { return 0 }

type pushSeg struct {
	data []*big.Int
	nrg  *big.Int
}

func (s pushSeg) do(program *Program, pc *uint64, env Environment, contract *Contract, memory *Memory, stack *stack) ([]byte, error) {
	// Use the calculated nrg. When insufficient nrg is present, use all nrg and return an
	// Out Of Nrg error
	if !contract.UseNrg(s.nrg) {
		return nil, OutOfNrgError
	}

	for _, d := range s.data {
		stack.push(new(big.Int).Set(d))
	}
	*pc += uint64(len(s.data))
	return nil, nil
}

func (s pushSeg) halts() bool { return false }
func (s pushSeg) Op() OpCode  { return 0 }
