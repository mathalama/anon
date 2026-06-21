const createBeep = (freq: number, duration: number, type: OscillatorType = 'sine') => {
  try {
    const AudioContext = window.AudioContext || (window as any).webkitAudioContext;
    if (!AudioContext) return;
    const ctx = new AudioContext();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    
    osc.type = type;
    osc.frequency.setValueAtTime(freq, ctx.currentTime);
    
    gain.gain.setValueAtTime(0.1, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.00001, ctx.currentTime + duration);
    
    osc.connect(gain);
    gain.connect(ctx.destination);
    
    osc.start();
    osc.stop(ctx.currentTime + duration);
  } catch (e) {
    console.error('AudioContext error:', e);
  }
};

export const playSound = (type: 'match' | 'message' | 'disconnect') => {
  if (typeof window === 'undefined') return;
  
  switch (type) {
    case 'match':
      createBeep(523.25, 0.4, 'sine'); // C5
      setTimeout(() => createBeep(659.25, 0.4, 'sine'), 100); // E5
      setTimeout(() => createBeep(783.99, 0.6, 'sine'), 200); // G5
      break;
    case 'message':
      createBeep(600, 0.2, 'sine');
      break;
    case 'disconnect':
      createBeep(300, 0.4, 'sawtooth');
      setTimeout(() => createBeep(250, 0.5, 'sawtooth'), 200);
      break;
  }
};
