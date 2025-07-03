import os
import shutil
import tempfile
import subprocess
from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.responses import PlainTextResponse

app = FastAPI()

EFLOMAL_ALIGN = "/app/eflomal/python/scripts/eflomal-align"

@app.post("/align", response_class=PlainTextResponse)
async def align_files(
    src: UploadFile = File(..., description="Source language file"),
    tgt: UploadFile = File(..., description="Target language file")
):
    # Use a temp dir for file operations
    with tempfile.TemporaryDirectory() as tmpdir:
        src_path = os.path.join(tmpdir, "src.txt")
        tgt_path = os.path.join(tmpdir, "tgt.txt")
        out_path = os.path.join(tmpdir, "out.align")
        
        # Save input files
        with open(src_path, "wb") as f:
            shutil.copyfileobj(src.file, f)
        with open(tgt_path, "wb") as f:
            shutil.copyfileobj(tgt.file, f)
        
        # Run eflomal-align
        try:
            result = subprocess.run(
                [
                    EFLOMAL_ALIGN,
                    "-s", src_path,
                    "-t", tgt_path,
                    "-f", out_path,
                ],
                check=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
        except subprocess.CalledProcessError as e:
            raise HTTPException(
                status_code=500,
                detail=f"eflomal-align failed: {e.stderr}"
            )
        
        # Return output file contents
        if not os.path.exists(out_path):
            raise HTTPException(status_code=500, detail="Alignment output missing.")
        with open(out_path, "r") as f:
            alignment = f.read()
        return alignment
