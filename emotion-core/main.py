import cv2
import time
from deepface import DeepFace
from utils.emotion_timer import EmotionTimer
import requests

# Inisialisasi
frames = []
frame_interval = 0.3  # detik
total_durasi = 5
jumlah_frame = int(total_durasi / frame_interval)

cap = cv2.VideoCapture(0)
print("Mulai rekam emosi selama 5 detik...")

# Ambil frame setiap 0.3 detik
for _ in range(jumlah_frame):
    ret, frame = cap.read()
    if ret:
        frames.append(frame)
    time.sleep(frame_interval)

cap.release()
cv2.destroyAllWindows()

# Deteksi emosi tiap frame
timer = EmotionTimer()
last_frame = None

print("Menganalisis emosi frame per frame...")

for frame in frames:
    try:
        result = DeepFace.analyze(frame, actions=['emotion'], enforce_detection=False)
        emosi_all = result[0]["emotion"]
        last_frame = frame
        timer.add_confidence(emosi_all, frame_interval)
        print("Emosi frame:", emosi_all)
    except Exception as e:
        print("Error:", str(e))
        continue

# Simpan frame terakhir
path_foto = f"snapshot_{int(time.time())}.jpg"
if last_frame is not None:
    cv2.imwrite(path_foto, last_frame)

# Hitung ringkasan
presentase, total = timer.get_summary()
emosi_dominan = max(presentase.items(), key=lambda x: x[1])[0]

durasi_netral = timer.data.get("Netral", 0)
durasi_bahagia = timer.data.get("Bahagia", 0)
durasi_tidak_bahagia = timer.data.get("Tidak Bahagia", 0)

print("Hasil Deteksi:")
print(f"- Total deteksi     : {round(total, 2)} detik")
print(f"- Netral            : {round(durasi_netral, 2)} detik ({presentase.get('Netral', 0)}%)")
print(f"- Bahagia           : {round(durasi_bahagia, 2)} detik ({presentase.get('Bahagia', 0)}%)")
print(f"- Tidak Bahagia     : {round(durasi_tidak_bahagia, 2)} detik ({presentase.get('Tidak Bahagia', 0)}%)")
print(f"Dominan             : {emosi_dominan}")
print(f"Gambar disimpan ke  : {path_foto}")

# Kirim ke backend
res = requests.post("http://localhost:8080/riwayat", files={
    "foto": open(path_foto, "rb")
}, data={
    "durasi_deteksi": total,
    "durasi_netral": durasi_netral,
    "durasi_bahagia": durasi_bahagia,
    "durasi_tidak_bahagia": durasi_tidak_bahagia,
    "presentase_netral": presentase.get("Netral", 0),
    "presentase_bahagia": presentase.get("Bahagia", 0),
    "presentase_tidak_bahagia": presentase.get("Tidak Bahagia", 0),
    "emosi_dominan": emosi_dominan
})

print("Hasil dikirim ke backend:")
print(f"Status: {res.status_code}")
print("Response:", res.text)
