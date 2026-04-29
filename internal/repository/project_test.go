package repository_test

import (
	"testing"

	"github.com/anjakDev/hourglass/internal/db"
	"github.com/anjakDev/hourglass/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) *repository.ProjectRepo {
	t.Helper()
	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return repository.NewProjectRepo(conn)
}

func TestProjectRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	id, err := repo.Create("Alpha", "first project", "#FF0000")
	require.NoError(t, err)
	assert.Positive(t, id)
}

func TestProjectRepo_Create_DuplicateName(t *testing.T) {
	repo := newTestDB(t)
	_, err := repo.Create("Alpha", "", "")
	require.NoError(t, err)
	_, err = repo.Create("Alpha", "", "")
	assert.Error(t, err, "duplicate name should fail")
}

func TestProjectRepo_GetByID(t *testing.T) {
	repo := newTestDB(t)
	id, err := repo.Create("Beta", "desc", "#00FF00")
	require.NoError(t, err)

	p, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, p.ID)
	assert.Equal(t, "Beta", p.Name)
	assert.Equal(t, "desc", p.Description)
	assert.Equal(t, "#00FF00", p.Color)
	assert.False(t, p.Archived)
}

func TestProjectRepo_GetByID_NotFound(t *testing.T) {
	repo := newTestDB(t)
	_, err := repo.GetByID(9999)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestProjectRepo_List_ExcludesArchived(t *testing.T) {
	repo := newTestDB(t)
	id1, _ := repo.Create("Active", "", "")
	id2, _ := repo.Create("Archived", "", "")
	require.NoError(t, repo.Archive(id2))

	projects, err := repo.List()
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, id1, projects[0].ID)
}

func TestProjectRepo_ListAll_IncludesArchived(t *testing.T) {
	repo := newTestDB(t)
	repo.Create("Active", "", "")
	id2, _ := repo.Create("Archived", "", "")
	require.NoError(t, repo.Archive(id2))

	projects, err := repo.ListAll()
	require.NoError(t, err)
	assert.Len(t, projects, 2)
}

func TestProjectRepo_List_Empty(t *testing.T) {
	repo := newTestDB(t)
	projects, err := repo.List()
	require.NoError(t, err)
	assert.Empty(t, projects)
}

func TestProjectRepo_List_OrderedByName(t *testing.T) {
	repo := newTestDB(t)
	repo.Create("Zebra", "", "")
	repo.Create("Apple", "", "")
	repo.Create("Mango", "", "")

	projects, err := repo.List()
	require.NoError(t, err)
	require.Len(t, projects, 3)
	assert.Equal(t, "Apple", projects[0].Name)
	assert.Equal(t, "Mango", projects[1].Name)
	assert.Equal(t, "Zebra", projects[2].Name)
}

func TestProjectRepo_Archive(t *testing.T) {
	repo := newTestDB(t)
	id, _ := repo.Create("ToArchive", "", "")
	require.NoError(t, repo.Archive(id))

	p, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.True(t, p.Archived)
}

func TestProjectRepo_Archive_NotFound(t *testing.T) {
	repo := newTestDB(t)
	err := repo.Archive(9999)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestProjectRepo_Delete(t *testing.T) {
	repo := newTestDB(t)
	id, _ := repo.Create("ToDelete", "", "")
	require.NoError(t, repo.Delete(id))

	_, err := repo.GetByID(id)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestProjectRepo_Delete_NotFound(t *testing.T) {
	repo := newTestDB(t)
	err := repo.Delete(9999)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
