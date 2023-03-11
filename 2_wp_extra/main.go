package main

import (
	"fmt"
	"context"
	"math/rand"
	"time"
)

type Task struct {
	Func func() error
}

/*
Worker -- выполняет задачи типа Tasks, досрочно завершается по команде
*/
type Worker struct {
	ctx context.Context
	stopCh chan struct{}
	tasksCh <-chan Task
}

func NewWorker(ctx context.Context, tasksCh <-chan Task) *Worker {
	return &Worker{
		ctx: ctx,
		stopCh: make(chan struct{}, 1),
		tasksCh: tasksCh,
	}
}

func (w *Worker) Run() {
	for {
		select {
			case <-w.ctx.Done():
				fmt.Println("Finishing worker: ctx done")
				return
			case <-w.stopCh:
				fmt.Println("Finishing worker: stop command")
				return
			case task := <-w.tasksCh:
				if err := task.Func(); err != nil {
					fmt.Printf("Worker: task finished with err %e\n", err)
				}
		}
		time.Sleep(time.Second)
	}
}

/*
Pool -- пул воркеров, увеличивает / уменьшает кол-во воркеров по команде
*/
type Pool struct {
	ctx context.Context
	addWorkerCh chan struct{}
	delWorkerCh chan struct{}
	tasksCh chan Task
	workers []*Worker
}

func NewPool(ctx context.Context) *Pool {
	pool := &Pool{
		ctx: ctx,
		addWorkerCh: make(chan struct{}, 1),
		delWorkerCh: make(chan struct{}, 1),
		tasksCh: make(chan Task, 2),
	}
	pool.AddWorker()
	return pool
}

func (p *Pool) AddWorker() {
	p.workers = append(p.workers, NewWorker(p.ctx, p.tasksCh))
	go p.workers[len(p.workers)-1].Run()
	fmt.Printf("Pool: AddWorker: num of workers: %d\n", len(p.workers))
}

func (p *Pool) DelWorker() {
	if len(p.workers) == 0 {
		return
	}
	p.workers[len(p.workers)-1].stopCh <- struct{}{}
	p.workers = p.workers[:len(p.workers)-1]
	fmt.Printf("Pool: AddWorker: num of workers: %d\n", len(p.workers))
}

func (p *Pool) Run() {
	for {
		select {
			case <-p.ctx.Done():
				fmt.Println("Finishing pool: ctx done")
				return
			case <-p.addWorkerCh:
				fmt.Println("Pool: AddWorker signal received")
				p.AddWorker()
			case <-p.delWorkerCh:
				fmt.Println("Pool: DelWorkerCh signal received")
				p.DelWorker()
		}
	}
}

/*
ResourceController -- периодически отправляет указания в pool через ctrlCh
логика определения команды может быть сложная, здесь просто рандом и sleep
*/
type ResourceController struct {
	ctx context.Context
	addWorkerCh chan<- struct{}
	delWorkerCh chan<- struct{}
}

func NewResourceController(ctx context.Context, addWorkerCh chan<- struct{}, delWorkerCh chan<- struct{}) *ResourceController {
	return &ResourceController{
		ctx: ctx,
		addWorkerCh: addWorkerCh,
		delWorkerCh: delWorkerCh,
	}
}

func (rc *ResourceController) Run() {
	for {
		select {
			case <-rc.ctx.Done():
				fmt.Println("Finishing ResourceController: ctx done")
				return
			default:
				dice := rand.Int()
				if dice % 3 != 0 {
					fmt.Println("ResourceController: AddWorker signal sent")
					go func() {
						rc.addWorkerCh<- struct{}{}
					} ()
				} else {
					fmt.Println("ResourceController: DelWorkerCh signal sent")
					go func() {
						rc.delWorkerCh<- struct{}{}
					} ()
				}
		}
		time.Sleep(time.Second)
	}
}

const UserCreatedTasksCnt = 10

func main() {
	pool := NewPool(context.Background())
	go pool.Run()
	
	ctrl := NewResourceController(context.Background(), pool.addWorkerCh, pool.delWorkerCh)
	go ctrl.Run()
	
	for i := 0; i < UserCreatedTasksCnt; i++ {
		fmt.Println("New user task crated")
		i := i
		go func() {
			pool.tasksCh <- Task{
				func() error {
					fmt.Printf("hello world from %d\n", i);
					return nil
				},
			}
		}()
	}
	time.Sleep(10 * time.Second)
}
