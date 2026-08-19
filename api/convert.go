package handler

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/restartfu/portmypack/portmypack"
	"github.com/restartfu/portmypack/portmypack/java"
)

// Handler is the Vercel Go serverless entrypoint for POST /api/convert.
// It accepts a multipart/form-data upload with a field named "pack"
// containing a Java Edition resource pack .zip, converts it to Bedrock
// (.mcpack) format, and streams the result back for download.
func Handler(w http.ResponseWriter, r *http.Request) {
	// CORS / preflight (harmless to keep even for same-origin use).
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32MB kept in memory; anything larger spills to os.TempDir() (which is
	// /tmp on Vercel's Go runtime, and is writable within a single invocation).
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "could not parse upload: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("pack")
	if err != nil {
		http.Error(w, "missing \"pack\" file field: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.EqualFold(filepath.Ext(header.Filename), ".zip") {
		http.Error(w, "please upload a .zip Java resource pack", http.StatusBadRequest)
		return
	}

	// Persist the upload to /tmp — the java package reads packs from a file path.
	inFile, err := os.CreateTemp("", "portmypack-in-*.zip")
	if err != nil {
		http.Error(w, "server error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	inPath := inFile.Name()
	defer os.Remove(inPath)

	if _, err := io.Copy(inFile, file); err != nil {
		inFile.Close()
		http.Error(w, "server error saving upload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	inFile.Close()

	javaPack, err := java.NewResourcePack(inPath)
	if err != nil {
		http.Error(w, "failed to read resource pack: "+err.Error(), http.StatusBadRequest)
		return
	}

	outPath := filepath.Join(os.TempDir(), fmt.Sprintf("portmypack-out-%d.mcpack", rand.IntN(999999)))
	defer os.Remove(outPath)

	portmypack.PortJavaEditionPack(javaPack, outPath)

	outFile, err := os.Open(outPath)
	if err != nil {
		http.Error(w, "server error reading result: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	base := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	downloadName := base + ".mcpack"

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
	if _, err := io.Copy(w, outFile); err != nil {
		// Response likely already partially written; nothing more we can do.
		return
	}
}
