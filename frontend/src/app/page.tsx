import Link from 'next/link';
import { MessageSquare, Mic, Shield, Zap, ArrowRight, Users } from 'lucide-react';

export default function LandingPage() {
  return (
    <div className="min-h-screen text-foreground selection:bg-primary/30 overflow-x-hidden">
      
      {/* Навигация */}
      <header className="fixed top-0 w-full z-50 glass">
        <div className="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3 font-bold text-xl tracking-tight group cursor-pointer">
            <div className="w-9 h-9 bg-gradient-to-br from-primary to-accent rounded-xl flex items-center justify-center shadow-lg shadow-primary/20 group-hover:rotate-12 transition-transform duration-300">
              <MessageSquare className="w-5 h-5 text-white" />
            </div>
            <span className="bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70">NektoKZ</span>
          </div>
          <nav>
            <Link 
              href="/search" 
              className="px-5 py-2 text-sm font-semibold glass rounded-full hover:bg-primary hover:text-white transition-all duration-300 active:scale-95"
            >
              Начать чат
            </Link>
          </nav>
        </div>
      </header>

      <main className="pt-40 pb-20 px-6 relative">
        <div className="max-w-6xl mx-auto">
          
          {/* Главный блок (Hero) */}
          <section className="text-center max-w-4xl mx-auto mb-32 relative">
            {/* Background Glow */}
            <div className="absolute -top-24 left-1/2 -translate-x-1/2 w-96 h-96 bg-primary/20 rounded-full blur-[120px] pointer-events-none" />
            
            <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass text-primary text-sm font-semibold mb-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
              <Zap className="w-4 h-4 fill-primary" />
              <span>Абсолютно анонимно • Без регистрации</span>
            </div>
            
            <h1 className="text-6xl md:text-8xl font-black tracking-tighter mb-8 leading-[0.9] animate-in fade-in slide-in-from-bottom-8 duration-1000">
              Общайся свободно. <br/>
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary via-accent to-primary bg-[length:200%_auto] animate-gradient-flow">
                С кем угодно.
              </span>
            </h1>
            
            <p className="text-xl md:text-2xl text-foreground/60 mb-12 max-w-2xl mx-auto leading-relaxed font-medium animate-in fade-in slide-in-from-bottom-12 duration-1000">
              Быстрые анонимные текстовые и голосовые чаты. Находи собеседников по интересам за секунды.
            </p>
            
            <div className="flex flex-col sm:flex-row items-center justify-center gap-6 animate-in fade-in slide-in-from-bottom-16 duration-1000">
              <Link 
                href="/search" 
                className="w-full sm:w-auto px-10 py-5 bg-gradient-to-br from-primary to-accent text-white rounded-2xl font-bold transition-all shadow-xl shadow-primary/30 flex items-center justify-center gap-3 group hover:scale-105 hover:shadow-primary/40 active:scale-95"
              >
                Найти собеседника
                <ArrowRight className="w-6 h-6 group-hover:translate-x-1 transition-transform" />
              </Link>
              <Link 
                href="#features" 
                className="w-full sm:w-auto px-10 py-5 glass hover:bg-foreground/5 rounded-2xl font-bold transition-all flex items-center justify-center active:scale-95"
              >
                Узнать больше
              </Link>
            </div>
          </section>

          {/* Преимущества (Features) */}
          <section id="features" className="scroll-mt-32">
            <div className="grid md:grid-cols-3 gap-8">
              
              {/* Карточка 1 */}
              <div className="group p-10 rounded-[2.5rem] glass hover:border-primary/50 transition-all duration-500 hover:-translate-y-2">
                <div className="w-14 h-14 bg-primary/10 text-primary rounded-2xl flex items-center justify-center mb-8 group-hover:scale-110 transition-transform duration-500">
                  <Shield className="w-7 h-7" />
                </div>
                <h3 className="text-2xl font-bold mb-4">Полная анонимность</h3>
                <p className="text-foreground/60 leading-relaxed font-medium">
                  Мы не требуем почту, телефон или имя. Зашел, выбрал настройки и сразу начал общаться.
                </p>
              </div>

              {/* Карточка 2 */}
              <div className="group p-10 rounded-[2.5rem] glass hover:border-accent/50 transition-all duration-500 hover:-translate-y-2">
                <div className="w-14 h-14 bg-accent/10 text-accent rounded-2xl flex items-center justify-center mb-8 group-hover:scale-110 transition-transform duration-500">
                  <Mic className="w-7 h-7" />
                </div>
                <h3 className="text-2xl font-bold mb-4">Текст и Голос</h3>
                <p className="text-foreground/60 leading-relaxed font-medium">
                  Выбирайте удобный формат. Хотите переписываться в тишине или пообщаться вживую?
                </p>
              </div>

              {/* Карточка 3 */}
              <div className="group p-10 rounded-[2.5rem] glass hover:border-emerald-500/50 transition-all duration-500 hover:-translate-y-2">
                <div className="w-14 h-14 bg-emerald-500/10 text-emerald-500 rounded-2xl flex items-center justify-center mb-8 group-hover:scale-110 transition-transform duration-500">
                  <Users className="w-7 h-7" />
                </div>
                <h3 className="text-2xl font-bold mb-4">Умный подбор</h3>
                <p className="text-foreground/60 leading-relaxed font-medium">
                  Настраивайте фильтры: указывайте свой пол, кого ищете и общие интересы.
                </p>
              </div>

            </div>
          </section>

        </div>
      </main>

      {/* Подвал */}
      <footer className="border-t border-border mt-32 glass">
        <div className="max-w-6xl mx-auto px-6 py-12 flex flex-col md:flex-row items-center justify-between text-sm text-foreground/50 font-medium">
          <p>© {new Date().getFullYear()} NektoKZ. Создано для свободного общения.</p>
          <div className="flex gap-8 mt-6 md:mt-0">
            <Link href="#" className="hover:text-primary transition-colors">Правила</Link>
            <Link href="#" className="hover:text-primary transition-colors">Конфиденциальность</Link>
          </div>
        </div>
      </footer>
    </div>

  );
}