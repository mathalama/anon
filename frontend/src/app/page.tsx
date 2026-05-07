import Link from 'next/link';
import { MessageSquare, Mic, Shield, Zap, ArrowRight, Users } from 'lucide-react';

export default function LandingPage() {
  return (
    <div className="min-h-screen text-white bg-[#050505] selection:bg-blue-500/30 overflow-x-hidden">
      
      {/* Навигация */}
      <header className="fixed top-0 w-full z-50 border-b border-white/5 bg-[#050505]/80 backdrop-blur-md">
        <div className="max-w-6xl mx-auto px-6 h-20 flex items-center justify-between">
          <div className="flex items-center gap-3 font-semibold text-xl tracking-tight group cursor-pointer">
            <div className="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center transition-transform group-hover:scale-105">
              <MessageSquare className="w-5 h-5 text-white fill-white/20" />
            </div>
            <span>NektoKZ</span>
          </div>
          <nav>
            <Link 
              href="/search" 
              className="sleek-button text-sm px-8 py-2.5"
            >
              Начать чат
            </Link>
          </nav>
        </div>
      </header>

      <main className="pt-48 pb-20 px-6">
        <div className="max-w-6xl mx-auto">
          
          {/* Главный блок (Hero) */}
          <section className="text-center max-w-4xl mx-auto mb-40">
            
            <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium mb-8">
              <Zap className="w-3.5 h-3.5 fill-blue-400" />
              <span>Абсолютно анонимно • Без регистрации</span>
            </div>
            
            <h1 className="text-6xl md:text-8xl font-bold tracking-tight mb-10 leading-[1.1]">
              Общайся <span className="bg-blue-600 px-3 py-1 text-white inline-block transform -rotate-2 rounded-lg">свободно.</span> <br/>
              <span className="text-zinc-500">С кем угодно.</span>
            </h1>
            
            <p className="text-xl md:text-2xl text-zinc-400 mb-12 max-w-2xl mx-auto leading-relaxed">
              Быстрые анонимные текстовые и голосовые чаты. <br className="hidden md:block" /> 
              Находи собеседников по интересам за секунды.
            </p>
            
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <Link 
                href="/search" 
                className="w-full sm:w-auto sleek-button text-lg py-4 px-10 flex items-center justify-center gap-3 group"
              >
                Найти собеседника
                <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
              </Link>
              <Link 
                href="#features" 
                className="w-full sm:w-auto sleek-button-secondary text-lg py-4 px-10"
              >
                Узнать больше
              </Link>
            </div>
          </section>

          {/* Преимущества (Features) */}
          <section id="features" className="scroll-mt-32">
            <div className="grid md:grid-cols-3 gap-6">
              
              <div className="sleek-card group hover:border-zinc-700 transition-colors">
                <div className="w-12 h-12 bg-zinc-900 border border-zinc-800 rounded-xl flex items-center justify-center mb-8 group-hover:bg-blue-600/10 group-hover:border-blue-600/20 transition-colors">
                  <Shield className="w-6 h-6 text-zinc-400 group-hover:text-blue-500 transition-colors" />
                </div>
                <h3 className="text-xl font-semibold mb-4">Полная анонимность</h3>
                <p className="text-zinc-500 leading-relaxed">
                  Мы не требуем почту, телефон или имя. Зашел, выбрал настройки и сразу начал общаться.
                </p>
              </div>

              <div className="sleek-card group hover:border-zinc-700 transition-colors">
                <div className="w-12 h-12 bg-zinc-900 border border-zinc-800 rounded-xl flex items-center justify-center mb-8 group-hover:bg-blue-600/10 group-hover:border-blue-600/20 transition-colors">
                  <Mic className="w-6 h-6 text-zinc-400 group-hover:text-blue-500 transition-colors" />
                </div>
                <h3 className="text-xl font-semibold mb-4">Текст и Голос</h3>
                <p className="text-zinc-500 leading-relaxed">
                  Выбирайте удобный формат. Хотите переписываться в тишине или пообщаться вживую?
                </p>
              </div>

              <div className="sleek-card group hover:border-zinc-700 transition-colors">
                <div className="w-12 h-12 bg-zinc-900 border border-zinc-800 rounded-xl flex items-center justify-center mb-8 group-hover:bg-blue-600/10 group-hover:border-blue-600/20 transition-colors">
                  <Users className="w-6 h-6 text-zinc-400 group-hover:text-blue-500 transition-colors" />
                </div>
                <h3 className="text-xl font-semibold mb-4">Умный подбор</h3>
                <p className="text-zinc-500 leading-relaxed">
                  Настраивайте фильтры: указывайте свой пол, кого ищете и общие интересы.
                </p>
              </div>

            </div>
          </section>

        </div>
      </main>

      {/* Подвал */}
      <footer className="border-t border-zinc-900 mt-40">
        <div className="max-w-6xl mx-auto px-6 py-12 flex flex-col md:flex-row items-center justify-between text-sm text-zinc-600">
          <p>© {new Date().getFullYear()} NektoKZ. Создано для свободного общения.</p>
          <div className="flex gap-8 mt-6 md:mt-0">
            <Link href="/rules" className="hover:text-white transition-colors">Правила</Link>
            <Link href="/privacy" className="hover:text-white transition-colors">Конфиденциальность</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
