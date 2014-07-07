package store

import ()

type BaseDirectorySPI interface {
	// DirectoryImplSPI
	LockID() string
}

/* Base implementation for a concrete Directory. */
type BaseDirectory struct {
	spi         BaseDirectorySPI
	IsOpen      bool
	lockFactory LockFactory
}

func NewBaseDirectory(spi BaseDirectorySPI) *BaseDirectory {
	assert(spi != nil)
	return &BaseDirectory{
		// DirectoryImpl: NewDirectoryImpl(spi),
		spi:    spi,
		IsOpen: true,
	}
}

func (d *BaseDirectory) MakeLock(name string) Lock {
	return d.lockFactory.Make(name)
}

func (d *BaseDirectory) ClearLock(name string) error {
	if d.lockFactory != nil {
		return d.lockFactory.Clear(name)
	}
	return nil
}

func (d *BaseDirectory) SetLockFactory(lockFactory LockFactory) {
	assert(d != nil && lockFactory != nil)
	d.lockFactory = lockFactory
	d.lockFactory.SetLockPrefix(d.spi.LockID())
}

func assert(ok bool) {
	if !ok {
		panic("assert fail")
	}
}

func (d *BaseDirectory) LockFactory() LockFactory {
	return d.lockFactory
}

func (d *BaseDirectory) EnsureOpen() {
	if !d.IsOpen {
		panic("this Directory is closed")
	}
}
