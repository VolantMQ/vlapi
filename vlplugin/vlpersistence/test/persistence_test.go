package test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/VolantMQ/vlapi/vlplugin/vlpersistence"
	persistenceMem "github.com/VolantMQ/vlapi/vlplugin/vlpersistence/mem"

	_ "github.com/VolantMQ/vlapi/vlplugin/vlpersistence/bbolt"
	_ "github.com/VolantMQ/vlapi/vlplugin/vlpersistence/mem"
)

var curr vlpersistence.IFace
var backends []vlpersistence.IFace

func init() {
	pl, _ := persistenceMem.Load(nil, nil)
	backends = append(backends, pl)
}

func TestMain(m *testing.M) {
	for _, f := range backends {
		curr = f
		m.Run()
	}
	os.Exit(0)
}

func TestSystemInfo(t *testing.T) {
	require.NotNil(t, curr)

	sys, err := curr.System()
	require.NoError(t, err)
	require.NotNil(t, sys)
}

func TestSessionCountZero(t *testing.T) {
	sessions, err := curr.Sessions()
	require.NoError(t, err)
	require.NotNil(t, sessions)

	count := sessions.Count()
	require.Equal(t, uint64(0), count)
}

func TestSessionDeleteNonExistingSession(t *testing.T) {
	sessions, err := curr.Sessions()
	require.NoError(t, err)
	require.NotNil(t, sessions)

	err = sessions.Delete([]byte("session"))
	require.EqualError(t, err, vlpersistence.ErrNotFound.Error())
}

func TestSessionCreateDelete(t *testing.T) {
	sessions, err := curr.Sessions()
	require.NoError(t, err)
	require.NotNil(t, sessions)

	err = sessions.Create([]byte("session"), &vlpersistence.SessionBase{
		Version:   1,
		Timestamp: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	count := sessions.Count()
	require.Equal(t, uint64(1), count)

	count, err = sessions.PacketCountQoS0([]byte("session"))
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)

	count, err = sessions.PacketCountQoS12([]byte("session"))
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)

	count, err = sessions.PacketCountUnAck([]byte("session"))
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)

	err = sessions.Delete([]byte("session"))
	require.NoError(t, err)

	err = sessions.Delete([]byte("session"))
	require.EqualError(t, err, vlpersistence.ErrNotFound.Error())
}

func TestSessionCreateDuplicate(t *testing.T) {
	sessions, err := curr.Sessions()
	require.NoError(t, err)
	require.NotNil(t, sessions)

	err = sessions.Create([]byte("session"), &vlpersistence.SessionBase{
		Version:   1,
		Timestamp: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

	err = sessions.Create([]byte("session"), &vlpersistence.SessionBase{
		Version:   1,
		Timestamp: time.Now().Format(time.RFC3339),
	})

	require.EqualError(t, err, vlpersistence.ErrAlreadyExists.Error())

	err = sessions.Delete([]byte("session"))
	require.NoError(t, err)
}

func TestSessionCreate(t *testing.T) {
	sessions, err := curr.Sessions()
	require.NoError(t, err)
	require.NotNil(t, sessions)

	err = sessions.Create([]byte("session"), &vlpersistence.SessionBase{
		Version:   1,
		Timestamp: time.Now().Format(time.RFC3339),
	})
	require.NoError(t, err)

}

func TestRetainedEmpty(t *testing.T) {
	retained, err := curr.Retained()
	require.NoError(t, err)
	require.NotNil(t, retained)

	packets, e := retained.Load()
	require.NoError(t, e)
	require.Len(t, packets, 0)
}

func TestRetained(t *testing.T) {
	packets := []*vlpersistence.PersistedPacket{
		{ExpireAt: time.Now().Format(time.RFC3339), Data: []byte{1, 2, 3, 4, 5}},
		{ExpireAt: time.Now().Format(time.RFC3339), Data: []byte{1, 2, 3, 4, 5}},
		{ExpireAt: time.Now().Format(time.RFC3339), Data: []byte{1, 2, 3, 4, 5}},
		{ExpireAt: time.Now().Format(time.RFC3339), Data: []byte{1, 2, 3, 4, 5}},
		{ExpireAt: time.Now().Format(time.RFC3339), Data: []byte{1, 2, 3, 4, 5}},
		{ExpireAt: time.Now().Format(time.RFC3339), Data: []byte{1, 2, 3, 4, 5}},
	}

	retained, err := curr.Retained()
	require.NoError(t, err)
	require.NotNil(t, retained)

	err = retained.Store(packets)
	require.NoError(t, err)

	storedPackets, e := retained.Load()
	require.NoError(t, e)
	require.Equal(t, packets, storedPackets)

	err = retained.Wipe()
	require.NoError(t, err)

	packets, e = retained.Load()
	require.NoError(t, e)
	require.Len(t, packets, 0)
}
