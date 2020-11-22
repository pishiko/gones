package apu

const sampleRate = 44100

type stream struct {
	remaining  []byte
	position   int64
	function   func(float64) float64
	IsActive   bool
	Time       float64
	Frequency  int
	volume     float64
	sweepCount int
}

func NewStream(function func(float64) float64, volume float64) *stream {
	str := &stream{function: function, volume: volume}
	return str
}

func (s *stream) Read(buf []byte) (int, error) {
	if s.IsActive && s.Time > 0 {
		if len(s.remaining) > 0 {
			n := copy(buf, s.remaining)
			s.remaining = s.remaining[n:]
			return n, nil
		}

		var origBuf []byte
		if len(buf)%4 > 0 {
			origBuf = buf
			buf = make([]byte, len(origBuf)+4-len(origBuf)%4)
		}
		var length = int64(sampleRate / s.Frequency)
		p := s.position / 4
		for i := 0; i < len(buf)/4; i++ {
			b := int16(s.function(float64(p)/float64(length)) * s.volume)
			buf[4*i] = byte(b)
			buf[4*i+1] = byte(b >> 8)
			buf[4*i+2] = byte(b)
			buf[4*i+3] = byte(b >> 8)
			p++
		}

		s.position += int64(len(buf))
		s.position %= length * 4

		if origBuf != nil {
			n := copy(origBuf, buf)
			s.remaining = buf[n:]
			return n, nil
		}
		return len(buf), nil

	} else {
		buf = make([]byte, len(buf))
		s.remaining = nil
		return len(buf), nil
	}
}

func (s *stream) Close() error {
	return nil
}

func squareWave0(x float64) float64 {
	if int(x/0.125)%8 == 0 {
		return 1
	} else {
		return -1
	}
}
func squareWave1(x float64) float64 {
	if int(x/0.25)%4 == 0 {
		return 1
	} else {
		return -1
	}
}
func squareWave2(x float64) float64 {
	if int(x/0.5)%2 == 0 {
		return 1
	} else {
		return -1
	}
}
func squareWave3(x float64) float64 {
	if int(x/0.25)%4 < 3 {
		return 1
	} else {
		return -1
	}
}
func triangleWave(x float64) float64 {
	n := int(x / 0.25)
	if n%2 == 0 {
		return 8*(x-float64(n)*0.25) - 1
	} else {

		return -8*(x-float64(n)*0.25) + 1
	}
}
func noiseWave(float64) float64 {
	return 0
}
