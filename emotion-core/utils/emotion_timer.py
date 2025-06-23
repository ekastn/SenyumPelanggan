from collections import defaultdict

class EmotionTimer:
    def __init__(self):
        self.data = defaultdict(float)
        self.total_duration = 0.0

    # Tambah berdasarkan label emosi dominan (1 emosi per frame)
    def add(self, emotion, interval):
        self.data[emotion] += interval
        self.total_duration += interval

    # Tambah berdasarkan confidence DeepFace
    def add_confidence(self, emotion_dict, interval):
        mapping = {
            "neutral": "Netral",
            "happy": "Bahagia",
            "sad": "Tidak Bahagia",
            "angry": "Tidak Bahagia",
            "disgust": "Tidak Bahagia",
            "fear": "Tidak Bahagia",
            "surprise": "Netral"
        }

        total_conf = sum(emotion_dict.values()) or 1
        for key, val in emotion_dict.items():
            mapped = mapping.get(key)
            if mapped:
                persen = val / total_conf
                self.data[mapped] += interval * persen

        self.total_duration += interval

    # Hitung persentase dan total
    def get_summary(self):
        total = self.total_duration
        if total == 0:
            return {emo: 0 for emo in self.data}, 0.0

        presentase = {
            emo: round((dur / total) * 100, 2)
            for emo, dur in self.data.items()
        }
        return presentase, round(total, 2)
