package claude

import (
	"bufio"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Subscriber receives parsed messages from a watched JSONL file.
type Subscriber struct {
	Ch   chan *RawMessage
	Done chan struct{}
}

// Watcher watches a single JSONL file for new lines and broadcasts parsed
// messages to all subscribers.
type Watcher struct {
	mu          sync.Mutex
	subscribers []*Subscriber
	path        string
	offset      int64
	fsw         *fsnotify.Watcher
	done        chan struct{}
}

// NewWatcher creates a new Watcher (does not start watching yet).
func NewWatcher() *Watcher {
	return &Watcher{
		done: make(chan struct{}),
	}
}

// Subscribe registers a new subscriber and returns it.
func (w *Watcher) Subscribe() *Subscriber {
	s := &Subscriber{
		Ch:   make(chan *RawMessage, 100),
		Done: make(chan struct{}),
	}
	w.mu.Lock()
	w.subscribers = append(w.subscribers, s)
	w.mu.Unlock()
	return s
}

// Unsubscribe removes a subscriber and closes its Done channel.
func (w *Watcher) Unsubscribe(s *Subscriber) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i, sub := range w.subscribers {
		if sub == s {
			w.subscribers = append(w.subscribers[:i], w.subscribers[i+1:]...)
			close(s.Done)
			return
		}
	}
}

// Watch starts watching the given JSONL file for appended lines.
// It seeks to the end of the file so only new content is delivered.
func (w *Watcher) Watch(path string) error {
	w.path = path

	// Seek to end of file to get initial offset.
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	w.offset = info.Size()

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := fsw.Add(path); err != nil {
		fsw.Close()
		return err
	}
	w.fsw = fsw

	go w.loop()
	return nil
}

// loop is the main goroutine that listens for fsnotify events.
func (w *Watcher) loop() {
	defer w.fsw.Close()
	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) {
				w.readNewLines()
			}
		case _, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
		}
	}
}

// readNewLines reads any bytes appended since the last offset, parses them as
// JSONL, and broadcasts renderable messages to all subscribers.
func (w *Watcher) readNewLines() {
	f, err := os.Open(w.path)
	if err != nil {
		return
	}
	defer f.Close()

	if _, err := f.Seek(w.offset, 0); err != nil {
		return
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		w.offset += int64(len(line)) + 1 // +1 for newline

		msg, err := ParseLine(line)
		if err != nil {
			continue
		}
		w.broadcast(msg)
	}
}

// broadcast sends a message to all current subscribers (non-blocking).
func (w *Watcher) broadcast(msg *RawMessage) {
	w.mu.Lock()
	subs := make([]*Subscriber, len(w.subscribers))
	copy(subs, w.subscribers)
	w.mu.Unlock()

	for _, s := range subs {
		select {
		case s.Ch <- msg:
		default:
			// Drop message if subscriber channel is full.
		}
	}
}

// Stop terminates the watcher goroutine and closes all subscriber Done channels.
func (w *Watcher) Stop() {
	select {
	case <-w.done:
		return // Already stopped.
	default:
		close(w.done)
	}

	w.mu.Lock()
	for _, s := range w.subscribers {
		select {
		case <-s.Done:
		default:
			close(s.Done)
		}
	}
	w.subscribers = nil
	w.mu.Unlock()
}

// ReadHistory reads all renderable messages from a JSONL file. If afterUUID is
// non-empty, only messages occurring after that UUID are returned.
func (w *Watcher) ReadHistory(path string, afterUUID string) ([]*RawMessage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	var msgs []*RawMessage
	found := afterUUID == ""

	for scanner.Scan() {
		msg, err := ParseLine(scanner.Bytes())
		if err != nil {
			continue
		}

		if !found {
			if msg.UUID == afterUUID {
				found = true
			}
			continue
		}

		if msg.IsRenderable() {
			msgs = append(msgs, msg)
		}
	}

	if err := scanner.Err(); err != nil {
		return msgs, err
	}
	return msgs, nil
}

// WatcherManager manages per-session Watchers, creating them on demand.
type WatcherManager struct {
	mu       sync.Mutex
	watchers map[string]*Watcher
}

// NewWatcherManager creates a new WatcherManager.
func NewWatcherManager() *WatcherManager {
	return &WatcherManager{
		watchers: make(map[string]*Watcher),
	}
}

// Get returns the Watcher for a session, creating and starting one if needed.
func (m *WatcherManager) Get(sessionID, jsonlPath string) (*Watcher, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if w, ok := m.watchers[sessionID]; ok {
		return w, nil
	}

	w := NewWatcher()
	if err := w.Watch(jsonlPath); err != nil {
		return nil, err
	}
	m.watchers[sessionID] = w
	return w, nil
}

// Remove stops and removes the Watcher for a session.
func (m *WatcherManager) Remove(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if w, ok := m.watchers[sessionID]; ok {
		w.Stop()
		delete(m.watchers, sessionID)
	}
}
