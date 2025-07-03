package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

var cmd *exec.Cmd

// convertToEflomalFormat converts a text file to eflomal binary format
func convertToEflomalFormat(input io.Reader, outputPath string) error {
    scanner := bufio.NewScanner(input)
    var sentences [][]string
    vocab := make(map[string]int)
    vocabIndex := 0
    
    // Read all sentences and build vocabulary
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        
        // Handle empty lines by creating empty sentences
        if line == "" {
            sentences = append(sentences, []string{})
            continue
        }
        
        tokens := strings.Fields(strings.ToLower(line))
        sentence := make([]string, len(tokens))
        
        for i, token := range tokens {
            sentence[i] = token
            if _, exists := vocab[token]; !exists {
                vocab[token] = vocabIndex
                vocabIndex++
            }
        }
        sentences = append(sentences, sentence)
    }
    
    if err := scanner.Err(); err != nil {
        return err
    }
    
    // Write eflomal format file
    outFile, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer outFile.Close()
    
    // Write header: number of sentences and vocabulary size
    fmt.Fprintf(outFile, "%d %d\n", len(sentences), len(vocab))
    
    // Write sentences
    for _, sentence := range sentences {
        if len(sentence) == 0 {
            fmt.Fprintf(outFile, "0\n")
            continue
        }
        
        fmt.Fprintf(outFile, "%d", len(sentence))
        for _, token := range sentence {
            fmt.Fprintf(outFile, " %d", vocab[token])
        }
        fmt.Fprintf(outFile, "\n")
    }
    
    return nil
}

func alignHandler(w http.ResponseWriter, r *http.Request) {
    // Enable CORS
    w.Header().Set("Access-Control-Allow-Origin", "*")
    if r.Method == http.MethodOptions {
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.WriteHeader(http.StatusOK)
        return
    }

    // Parse multipart form
    err := r.ParseMultipartForm(32 << 20)
    if err != nil {
        http.Error(w, "Invalid form", http.StatusBadRequest)
        return
    }

    src, _, err := r.FormFile("src")
    if err != nil {
        http.Error(w, "Missing src file", http.StatusBadRequest)
        return
    }
    defer src.Close()

    tgt, _, err := r.FormFile("tgt")
    if err != nil {
        http.Error(w, "Missing tgt file", http.StatusBadRequest)
        return
    }
    defer tgt.Close()

    tmpdir, err := os.MkdirTemp("", "align")
    if err != nil {
        http.Error(w, "Temp dir error", http.StatusInternalServerError)
        return
    }
    // defer os.RemoveAll(tmpdir)

    srcPath := filepath.Join(tmpdir, "src.txt")
    tgtPath := filepath.Join(tmpdir, "tgt.txt")
    fwdPath := filepath.Join(tmpdir, "out.fwd")
    revPath := filepath.Join(tmpdir, "out.rev")
    symPath := filepath.Join(tmpdir, "out.sym")

    // Save files in eflomal format
    err = convertToEflomalFormat(src, srcPath)
    if err != nil {
        http.Error(w, "Failed to convert src file: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    err = convertToEflomalFormat(tgt, tgtPath)
    if err != nil {
        http.Error(w, "Failed to convert tgt file: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Run eflomal-align FORWARD
    cmd = exec.Command("/app/eflomal", "-s", srcPath, "-t", tgtPath, "-f", fwdPath, "-r", revPath, "-m", "3")
    if out, err := cmd.CombinedOutput(); err != nil {
        http.Error(w, "eflomal-align failed: "+string(out), 500)
        return
    }

    // Symmetrize with atools
    cmd = exec.Command("atools", "-i", fwdPath, "-j", revPath, "-c", "grow-diag-final-and")
    symFile, err := os.Create(symPath)
    if err != nil {
        http.Error(w, "Failed to create sym file: "+err.Error(), http.StatusInternalServerError)
        return
    }
    cmd.Stdout = symFile
    cmd.Stderr = symFile // Capture stderr as well
    
    if err := cmd.Run(); err != nil {
        symFile.Close()
        // Read the error output from the file
        if errorData, readErr := os.ReadFile(symPath); readErr == nil {
            http.Error(w, "atools failed: "+string(errorData), 500)
        } else {
            http.Error(w, "atools failed: "+err.Error(), 500)
        }
        return
    }
    symFile.Close()

    // Return result
    symData, err := os.ReadFile(symPath)
    if err != nil {
        http.Error(w, "Failed to read output", 500)
        return
    }
    w.Header().Set("Content-Type", "text/plain")
    w.Write(symData)
}

func main() {
    http.HandleFunc("/align", alignHandler)
    log.Fatal(http.ListenAndServe(":8000", nil))
}