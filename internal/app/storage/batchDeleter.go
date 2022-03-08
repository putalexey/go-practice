package storage

import (
	"context"
	"log"
	"sync"
	"time"
)

var deleteQueryTimeout = 30 * time.Second

type DeleteTask struct {
	shorts []string
	userID string
}

type BatchDeleter struct {
	//addMutex   *sync.Mutex
	flushMutex *sync.Mutex
	cond       *sync.Cond
	store      Storager
	//inputChan     []DeleteTask
	inputChan  chan DeleteTask
	bufferSize int
	ticker     *time.Ticker
	ctx        context.Context
}

func NewBatchDeleterWithContext(ctx context.Context, store Storager, bufferSize int) *BatchDeleter {
	//ctx, cancel := context.WithCancel(context.Background())
	fm := &sync.Mutex{}
	deleter := BatchDeleter{
		store:      store,
		flushMutex: fm,
		cond:       sync.NewCond(fm),
		bufferSize: bufferSize,
		inputChan:  make(chan DeleteTask, bufferSize),
		ticker:     time.NewTicker(10 * time.Second),
		ctx:        ctx,
	}

	go deleter.Start()

	return &deleter
}

func (b *BatchDeleter) QueueItems(shorts []string, userID string) {
	go func() {
		log.Println("DEBUG: adding to queue", shorts)
		b.inputChan <- DeleteTask{
			shorts: shorts,
			userID: userID,
		}

		// if channel is full => call signal earlier than timer
		if len(b.inputChan) == cap(b.inputChan) {
			b.cond.Signal()
		}
	}()
}

func (b *BatchDeleter) flushWorker() {
	for {
		tasksQueue := b.doWork()

		if len(tasksQueue) > 0 {
			ctx, cancel := context.WithTimeout(b.ctx, deleteQueryTimeout)
			for userID, t := range tasksQueue {
				records, err := b.store.LoadBatch(ctx, t.shorts)
				if err != nil {
					log.Println("WARNING: ", err)
					continue
				}
				for _, r := range records {
					if r.UserID != userID {
						log.Println("WARNING: ", userID, " can't delete item ", r.Short, err)
						continue
					}
				}

				if err := b.store.DeleteBatch(ctx, t.shorts); err != nil {
					log.Println("WARNING: ", err)
					continue
				}
				log.Println("DEBUG: flushed", t)
			}
			cancel()
		} else {
			log.Println("DEBUG: nothing to flush")
		}
	}
}

func (b *BatchDeleter) doWork() (queue map[string]*DeleteTask) {
	b.flushMutex.Lock()
	b.cond.Wait()
	defer b.flushMutex.Unlock()

	// queue is map of delete tasks merged by userid
	queue = make(map[string]*DeleteTask)
	for {
		select {
		case t, ok := <-b.inputChan:
			if ok {
				if v, ok := queue[t.userID]; ok {
					v.shorts = append(v.shorts, t.shorts...)
				} else {
					queue[t.userID] = &DeleteTask{t.shorts, t.userID}
				}
				log.Println("DEBUG: adding to flush", t.shorts)
			} else {
				// Channel closed!
				return
			}
		default:
			// channel is empty, wait for next signal
			return
		}
	}
}

func (b *BatchDeleter) Start() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		b.flushWorker()
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-b.ticker.C:
				b.cond.Signal()
			case <-b.ctx.Done():
				b.ticker.Stop()
				return
			}
		}
	}()

	wg.Wait()
}
