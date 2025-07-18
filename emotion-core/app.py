# emotion-core/app.py
from flask import Flask, request, jsonify
import sys
import json
import base64
import io
import cv2
import time
import numpy as np
from deepface import DeepFace
from utils.emotion_timer import EmotionTimer
import os

app = Flask(__name__)

def map_emosi(emosi):
    if emosi == "neutral":
        return "Netral"
    elif emosi == "happy":
        return "Bahagia"
    else:
        return "Tidak Bahagia"

@app.route('/detect', methods=['POST'])
def detect_emotion():
    data = request.json
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
            # print(f"Error processing frame: {e}", file=sys.stderr)
            continue

    # Simpan gambar terakhir dan encode ke base64
    encoded_image = ""
    if last_frame is not None:
        _, buffer = cv2.imencode('.jpg', last_frame)
        encoded_image = base64.b64encode(buffer).decode('utf-8')

    presentase, total = timer.get_summary()
    emosi_dominan = max(presentase.items(), key=lambda x: x[1])[0]
    durasi_netral = timer.data.get("Netral", 0)
    durasi_bahagia = timer.data.get("Bahagia", 0)
    durasi_tidak_bahagia = timer.data.get("Tidak Bahagia", 0)

    return jsonify({
        "durasi_deteksi": total,
        "durasi_netral": durasi_netral,
        "durasi_bahagia": durasi_bahagia,
        "durasi_tidak_bahagia": durasi_tidak_bahagia,
        "presentase_netral": presentase.get("Netral", 0),
        "presentase_bahagia": presentase.get("Bahagia", 0),
        "presentase_tidak_bahagia": presentase.get("Tidak Bahagia", 0),
        "emosi_dominan": emosi_dominan,
        "foto_base64": encoded_image
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)