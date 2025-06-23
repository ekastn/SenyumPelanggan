# emotion-core/deteksi_batch.py
import sys
import json
import base64
import io
import cv2
import time
import numpy as np
from deepface import DeepFace
from utils.emotion_timer import EmotionTimer

def map_emosi(emosi):
    if emosi == "neutral":
        return "Netral"
    elif emosi == "happy":
        return "Bahagia"
    else:
        return "Tidak Bahagia"

# Terima input dari stdin
payload = sys.stdin.read()
data = json.loads(payload)
images = data.get("frames", [])
frame_interval = data.get("interval", 0.5)

timer = EmotionTimer()
last_frame = None

for img_b64 in images:
    try:
        img_bytes = base64.b64decode(img_b64)
        npimg = np.frombuffer(img_bytes, np.uint8)
        frame = cv2.imdecode(npimg, cv2.IMREAD_COLOR)
        result = DeepFace.analyze(frame, actions=['emotion'], enforce_detection=False)
        emosi_asli = result[0]['dominant_emotion']
        kategori = map_emosi(emosi_asli)
        timer.add(kategori, frame_interval)
        last_frame = frame
    except Exception as e:
        continue

# Simpan gambar terakhir
path_foto = f"snapshot_{int(time.time())}.jpg"
if last_frame is not None:
    cv2.imwrite(path_foto, last_frame)

presentase, total = timer.get_summary()
emosi_dominan = max(presentase.items(), key=lambda x: x[1])[0]
durasi_netral = timer.data.get("Netral", 0)
durasi_bahagia = timer.data.get("Bahagia", 0)
durasi_tidak_bahagia = timer.data.get("Tidak Bahagia", 0)

# Output hasil ke stdout
print(json.dumps({
    "durasi_deteksi": total,
    "durasi_netral": durasi_netral,
    "durasi_bahagia": durasi_bahagia,
    "durasi_tidak_bahagia": durasi_tidak_bahagia,
    "presentase_netral": presentase.get("Netral", 0),
    "presentase_bahagia": presentase.get("Bahagia", 0),
    "presentase_tidak_bahagia": presentase.get("Tidak Bahagia", 0),
    "emosi_dominan": emosi_dominan,
    "path_foto": path_foto
}))
