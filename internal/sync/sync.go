package sync

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andragon31/fenrir/internal/graph"
)

type Sync struct {
	graph    *graph.Graph
	chunksDir string
}

type Chunk struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Nodes     []graph.Node `json:"nodes"`
	Edges     []graph.Edge `json:"edges"`
	Sessions  []graph.Session `json:"sessions"`
	Checksum  string    `json:"checksum"`
}

type Manifest struct {
	Version   int       `json:"version"`
	Chunks    []ChunkMeta `json:"chunks"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChunkMeta struct {
	ID        string    `json:"id"`
	Checksum  string    `json:"checksum"`
	Timestamp time.Time `json:"timestamp"`
	NodeCount int       `json:"node_count"`
	EdgeCount int       `json:"edge_count"`
}

func New(g *graph.Graph, chunksDir string) *Sync {
	return &Sync{
		graph:    g,
		chunksDir: chunksDir,
	}
}

func (s *Sync) Export() error {
	if err := os.MkdirAll(s.chunksDir, 0755); err != nil {
		return fmt.Errorf("failed to create chunks directory: %w", err)
	}

	nodes, err := s.exportNodes()
	if err != nil {
		return err
	}

	edges, err := s.exportEdges()
	if err != nil {
		return err
	}

	sessions, err := s.exportSessions()
	if err != nil {
		return err
	}

	chunk := Chunk{
		ID:        generateChunkID(),
		Timestamp: time.Now().UTC(),
		Nodes:     nodes,
		Edges:     edges,
		Sessions:  sessions,
	}

	content, err := json.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk: %w", err)
	}

	chunk.Checksum = s.calculateChecksum(content)

	data, err := json.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk with checksum: %w", err)
	}

	filename := chunk.Checksum + ".jsonl.gz"
	filepath := filepath.Join(s.chunksDir, filename)

	if err := s.writeGzippedJSONL(filepath, data); err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	if err := s.updateManifest(chunk); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}

func (s *Sync) exportNodes() ([]graph.Node, error) {
	nodes, _, err := s.graph.GetTimeline("", 10000)
	return nodes, err
}

func (s *Sync) exportEdges() ([]graph.Edge, error) {
	return []graph.Edge{}, nil
}

func (s *Sync) exportSessions() ([]graph.Session, error) {
	return s.graph.ListSessions(1000)
}

func (s *Sync) writeGzippedJSONL(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := gzip.NewWriter(file)
	defer writer.Close()

	_, err = writer.Write(data)
	return err
}

func (s *Sync) calculateChecksum(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

func (s *Sync) updateManifest(chunk Chunk) error {
	manifest, err := s.loadManifest()
	if err != nil {
		manifest = &Manifest{
			Version:   1,
			Chunks:    []ChunkMeta{},
			UpdatedAt: time.Now().UTC(),
		}
	}

	for i, existing := range manifest.Chunks {
		if existing.Checksum == chunk.Checksum {
			return nil
		}
		if existing.ID == chunk.ID {
			manifest.Chunks[i] = ChunkMeta{
				ID:        chunk.ID,
				Checksum:  chunk.Checksum,
				Timestamp: chunk.Timestamp,
				NodeCount: len(chunk.Nodes),
				EdgeCount: len(chunk.Edges),
			}
			manifest.UpdatedAt = time.Now().UTC()
			return s.saveManifest(manifest)
		}
	}

	manifest.Chunks = append(manifest.Chunks, ChunkMeta{
		ID:        chunk.ID,
		Checksum:  chunk.Checksum,
		Timestamp: chunk.Timestamp,
		NodeCount: len(chunk.Nodes),
		EdgeCount: len(chunk.Edges),
	})
	manifest.UpdatedAt = time.Now().UTC()

	return s.saveManifest(manifest)
}

func (s *Sync) loadManifest() (*Manifest, error) {
	manifestPath := filepath.Join(s.chunksDir, "manifest.json")
	
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (s *Sync) saveManifest(manifest *Manifest) error {
	manifestPath := filepath.Join(s.chunksDir, "manifest.json")
	
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(manifestPath, data, 0644)
}

func (s *Sync) Import() error {
	manifest, err := s.loadManifest()
	if err != nil {
		return fmt.Errorf("no manifest found: %w", err)
	}

	imported := 0
	for _, chunkMeta := range manifest.Chunks {
		filepath := filepath.Join(s.chunksDir, chunkMeta.Checksum+".jsonl.gz")
		
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			continue
		}

		chunk, err := s.readChunk(filepath)
		if err != nil {
			continue
		}

		if err := s.mergeChunk(chunk); err != nil {
			continue
		}

		imported++
	}

	return nil
}

func (s *Sync) readChunk(path string) (*Chunk, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var chunk Chunk
	if err := json.NewDecoder(reader).Decode(&chunk); err != nil {
		return nil, err
	}

	return &chunk, nil
}

func (s *Sync) mergeChunk(chunk *Chunk) error {
	for _, node := range chunk.Nodes {
		existing := s.nodeExists(node.ID)
		if existing {
			if chunk.Timestamp.After(existing.Timestamp) {
				s.graph.SaveNode(&node)
			}
		} else {
			s.graph.SaveNode(&node)
		}
	}

	for _, edge := range chunk.Edges {
		if !s.edgeExists(edge.ID) {
			s.graph.SaveEdge(&edge)
		}
	}

	for _, session := range chunk.Sessions {
		if !s.sessionExists(session.ID) {
			s.graph.StartSession(session.Goal, "")
		}
	}

	return nil
}

func (s *Sync) nodeExists(id string) (graph.Node, bool) {
	nodes, _, _ := s.graph.GetTimeline(id, 0)
	if len(nodes) > 0 {
		return nodes[0], true
	}
	return graph.Node{}, false
}

func (s *Sync) edgeExists(id string) bool {
	return false
}

func (s *Sync) sessionExists(id string) bool {
	sessions, _ := s.graph.ListSessions(10000)
	for _, s := range sessions {
		if s.ID == id {
			return true
		}
	}
	return false
}

func (s *Sync) GetStatus() (*graph.SyncStatus, error) {
	manifest, err := s.loadManifest()
	if err != nil {
		return &graph.SyncStatus{
			Chunks:   0,
			LastSync: "never",
		}, nil
	}

	status := &graph.SyncStatus{
		Chunks: len(manifest.Chunks),
	}

	if len(manifest.Chunks) > 0 {
		lastChunk := manifest.Chunks[len(manifest.Chunks)-1]
		status.LastSync = lastChunk.Timestamp.Format(time.RFC3339)
	}

	return status, nil
}

func generateChunkID() string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%d-%d", timestamp, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

func (s *Sync) ExportChunked(maxNodesPerChunk int) error {
	if err := os.MkdirAll(s.chunksDir, 0755); err != nil {
		return err
	}

	manifest := &Manifest{
		Version:   1,
		Chunks:    []ChunkMeta{},
		UpdatedAt: time.Now().UTC(),
	}

	allNodes, _, err := s.graph.GetTimeline("", 100000)
	if err != nil {
		return err
	}

	totalChunks := (len(allNodes) + maxNodesPerChunk - 1) / maxNodesPerChunk

	for i := 0; i < totalChunks; i++ {
		start := i * maxNodesPerChunk
		end := start + maxNodesPerChunk
		if end > len(allNodes) {
			end = len(allNodes)
		}

		chunkNodes := allNodes[start:end]

		chunk := Chunk{
			ID:        fmt.Sprintf("chunk-%d-of-%d", i+1, totalChunks),
			Timestamp: time.Now().UTC(),
			Nodes:     chunkNodes,
			Edges:     []graph.Edge{},
			Sessions:  []graph.Session{},
		}

		content, _ := json.Marshal(chunk)
		chunk.Checksum = s.calculateChecksum(content)

		filename := fmt.Sprintf("%s.jsonl.gz", chunk.Checksum)
		filepath := filepath.Join(s.chunksDir, filename)

		if err := s.writeGzippedJSONL(filepath, content); err != nil {
			return err
		}

		manifest.Chunks = append(manifest.Chunks, ChunkMeta{
			ID:        chunk.ID,
			Checksum:  chunk.Checksum,
			Timestamp: chunk.Timestamp,
			NodeCount: len(chunkNodes),
			EdgeCount: 0,
		})
	}

	return s.saveManifest(manifest)
}

func (s *Sync) MergeFromRemote(remotePath string) error {
	remoteSync := New(s.graph, remotePath)
	
	manifest, err := remoteSync.loadManifest()
	if err != nil {
		return err
	}

	merged := 0
	for _, chunkMeta := range manifest.Chunks {
		filepath := filepath.Join(remotePath, chunkMeta.Checksum+".jsonl.gz")
		
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			continue
		}

		chunk, err := remoteSync.readChunk(filepath)
		if err != nil {
			continue
		}

		if err := s.mergeChunk(chunk); err != nil {
			continue
		}

		merged++
	}

	return nil
}

func (s *Sync) ValidateChunks() ([]string, error) {
	manifest, err := s.loadManifest()
	if err != nil {
		return nil, err
	}

	var invalidChunks []string

	for _, chunkMeta := range manifest.Chunks {
		filepath := filepath.Join(s.chunksDir, chunkMeta.Checksum+".jsonl.gz")
		
		file, err := os.Open(filepath)
		if err != nil {
			invalidChunks = append(invalidChunks, chunkMeta.ID+": file not found")
			continue
		}

		reader, err := gzip.NewReader(file)
		if err != nil {
			file.Close()
			invalidChunks = append(invalidChunks, chunkMeta.ID+": invalid gzip format")
			continue
		}

		data, err := io.ReadAll(reader)
		reader.Close()
		file.Close()
		if err != nil {
			invalidChunks = append(invalidChunks, chunkMeta.ID+": failed to read")
			continue
		}

		calculatedChecksum := s.calculateChecksum(data)
		if calculatedChecksum != chunkMeta.Checksum {
			invalidChunks = append(invalidChunks, chunkMeta.ID+": checksum mismatch")
		}
	}

	return invalidChunks, nil
}

func (s *Sync) CleanOrphanedChunks() error {
	manifest, err := s.loadManifest()
	if err != nil {
		return err
	}

	validChecksums := make(map[string]bool)
	for _, chunk := range manifest.Chunks {
		validChecksums[chunk.Checksum] = true
	}

	entries, err := os.ReadDir(s.chunksDir)
	if err != nil {
		return err
	}

	cleaned := 0
	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".jsonl.gz") {
			checksum := strings.TrimSuffix(entry.Name(), ".jsonl.gz")
			if !validChecksums[checksum] && len(checksum) == 64 {
				filepath := filepath.Join(s.chunksDir, entry.Name())
				if err := os.Remove(filepath); err == nil {
					cleaned++
				}
			}
		}
	}

	if cleaned > 0 {
		fmt.Printf("Cleaned %d orphaned chunks\n", cleaned)
	}

	return nil
}
