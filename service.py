import os
import shutil
import tempfile
import subprocess
from fastapi import FastAPI, UploadFile, File, HTTPException
from fastapi.responses import PlainTextResponse

app = FastAPI()

EFLOMAL_ALIGN = "/app/eflomal/python/scripts/eflomal-align"
ATOOLS = "atools"  # Adjust path if atools is not in PATH

@app.post("/align", response_class=PlainTextResponse)
async def align_files(
    src: UploadFile = File(..., description="Source language file"),
    tgt: UploadFile = File(..., description="Target language file")
):
    # Use a temp dir for file operations
    with tempfile.TemporaryDirectory() as tmpdir:
        src_path = os.path.join(tmpdir, "src.txt")
        tgt_path = os.path.join(tmpdir, "tgt.txt")
        fwd_path = os.path.join(tmpdir, "out.fwd")
        rev_path = os.path.join(tmpdir, "out.rev")
        sym_path = os.path.join(tmpdir, "out.sym")

        # Save input files
        with open(src_path, "wb") as f:
            shutil.copyfileobj(src.file, f)
        with open(tgt_path, "wb") as f:
            shutil.copyfileobj(tgt.file, f)

        # Run eflomal-align FORWARD
        try:
            subprocess.run(
                [
                    EFLOMAL_ALIGN,
                    "-s", src_path,
                    "-t", tgt_path,
                    "-f", fwd_path,
                ],
                check=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
        except subprocess.CalledProcessError as e:
            raise HTTPException(
                status_code=500,
                detail=f"eflomal-align forward failed: {e.stderr}"
            )

        # Run eflomal-align REVERSE
        try:
            subprocess.run(
                [
                    EFLOMAL_ALIGN,
                    "-s", tgt_path,
                    "-t", src_path,
                    "-f", rev_path,
                ],
                check=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
        except subprocess.CalledProcessError as e:
            raise HTTPException(
                status_code=500,
                detail=f"eflomal-align reverse failed: {e.stderr}"
            )

        # Symmetrize with atools (grow-diag-final-and)
        try:
            subprocess.run(
                [
                    ATOOLS,
                    "-i", fwd_path,
                    "-j", rev_path,
                    "-c", "grow-diag-final-and"
                ],
                check=True,
                stdout=open(sym_path, "w"),
                stderr=subprocess.PIPE,
                text=True
            )
        except subprocess.CalledProcessError as e:
            raise HTTPException(
                status_code=500,
                detail=f"atools symmetrization failed: {e.stderr}"
            )

        # Return symmetrized output file contents
        if not os.path.exists(sym_path):
            raise HTTPException(status_code=500, detail="Symmetrized alignment output missing.")
        with open(sym_path, "r") as f:
            alignment = f.read()
        return alignment
