// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xcycle

import (
	"sync"
	"sync/atomic"
)

//Cycle ..
type Cycle struct {
	mu       *sync.Mutex
	wg       *sync.WaitGroup
	done     chan struct{}
	quit     chan error
	closeing uint32
	waiting  uint32
}

//NewCycle new a cycle life
func NewCycle() *Cycle {
	return &Cycle{
		mu:       &sync.Mutex{},
		wg:       &sync.WaitGroup{},
		done:     make(chan struct{}),
		quit:     make(chan error),
		closeing: 0,
		waiting:  0,
	}
}

//Run a new goroutine
func (c *Cycle) Run(fn func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.wg.Add(1)
	go func(c *Cycle) {
		defer c.wg.Done()
		if err := fn(); err != nil {
			c.quit <- err
		}
	}(c)
}

//Done block and return a chan error
func (c *Cycle) Done() <-chan struct{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	if atomic.CompareAndSwapUint32(&c.waiting, 0, 1) {
		go func() {
			c.wg.Wait()
			close(c.done)
		}()
	}
	return c.done
}

//DoneAndClose ..
func (c *Cycle) DoneAndClose() {
	<-c.Done()
	c.Close()
}

//Close ..
func (c *Cycle) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if atomic.CompareAndSwapUint32(&c.closeing, 0, 1) {
		close(c.quit)
	}
}

// Wait blocked for a life cycle
func (c *Cycle) Wait() <-chan error {
	return c.quit
}
