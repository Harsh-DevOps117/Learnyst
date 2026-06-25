'use client';

import { PrimaryNav } from '@/components/ui/PrimaryNav';
import Link from 'next/link';
import { useSearchParams, useRouter } from 'next/navigation';
import { Suspense, useState, useRef } from 'react';
import { Menu, X, LogOut } from 'lucide-react';

import gsap from 'gsap';
import { useGSAP } from '@gsap/react';

import { useAuthStore } from '@/store/authStore';

function NavLinks({ currentTab, onNavigate }: { currentTab: string, onNavigate?: () => void }) {
  const { roleId } = useAuthStore();

  const getLinkClass = (tab: string) => {
    return currentTab === tab 
      ? "bg-[#1ED760] text-black px-3 py-2 font-semibold uppercase tracking-widest text-[12px] block border-2 border-black"
      : "text-black px-3 py-2 font-semibold uppercase tracking-widest text-[12px] hover:bg-gray-200 transition-colors block border-2 border-transparent hover:border-black";
  };

  return (
    <>
      {roleId === 4 && (
        <>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">Overview</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=global_analytics" onClick={onNavigate} className={getLinkClass('global_analytics')}>Platform Analytics</Link></li>
            </ul>
          </div>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">Manage</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=all_users" onClick={onNavigate} className={getLinkClass('all_users')}>All Users</Link></li>
            </ul>
          </div>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">Account</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=profile" onClick={onNavigate} className={getLinkClass('profile')}>Profile</Link></li>
            </ul>
          </div>
        </>
      )}

      {roleId === 1 && (
        <>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">Overview</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=academy" onClick={onNavigate} className={getLinkClass('academy')}>My Academy</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=templates" onClick={onNavigate} className={getLinkClass('templates')}>Website Theme</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=performance" onClick={onNavigate} className={getLinkClass('performance')}>Performance</Link></li>
            </ul>
          </div>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">Manage</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=courses" onClick={onNavigate} className={getLinkClass('courses')}>Courses</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=teachers" onClick={onNavigate} className={getLinkClass('teachers')}>Teachers</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=students" onClick={onNavigate} className={getLinkClass('students')}>Students</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=subscriptions" onClick={onNavigate} className={getLinkClass('subscriptions')}>Subscriptions</Link></li>
            </ul>
          </div>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">Account</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=profile" onClick={onNavigate} className={getLinkClass('profile')}>Profile</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=settings" onClick={onNavigate} className={getLinkClass('settings')}>Settings</Link></li>
            </ul>
          </div>
        </>
      )}

      {roleId === 2 && (
        <>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">Teaching</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=courses" onClick={onNavigate} className={getLinkClass('courses')}>My Courses</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=assignments" onClick={onNavigate} className={getLinkClass('assignments')}>Assignments</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=students" onClick={onNavigate} className={getLinkClass('students')}>Students</Link></li>
            </ul>
          </div>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">Account</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=profile" onClick={onNavigate} className={getLinkClass('profile')}>Profile</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=settings" onClick={onNavigate} className={getLinkClass('settings')}>Payout Settings</Link></li>
            </ul>
          </div>
        </>
      )}

      {roleId === 3 && (
        <>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">My Learning</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=learning" onClick={onNavigate} className={getLinkClass('learning')}>Courses</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=assignments" onClick={onNavigate} className={getLinkClass('assignments')}>Assignments</Link></li>
              <li className="w-full"><Link href="/dashboard?tab=quizzes" onClick={onNavigate} className={getLinkClass('quizzes')}>Quizzes</Link></li>
            </ul>
          </div>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">Tools</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=ai_tutor" onClick={onNavigate} className={getLinkClass('ai_tutor')}>AI Tutor</Link></li>
            </ul>
          </div>
          <div>
            <h3 className="font-ui font-bold text-[16px] text-black mb-4 uppercase tracking-widest border-b-2 border-black pb-2">Account</h3>
            <ul className="space-y-3 flex flex-col items-start w-full">
              <li className="w-full"><Link href="/dashboard?tab=profile" onClick={onNavigate} className={getLinkClass('profile')}>Profile</Link></li>
            </ul>
          </div>
        </>
      )}
    </>
  );
}

function Sidebar() {
  const searchParams = useSearchParams();
  const { roleId, logout } = useAuthStore();
  const currentTab = searchParams?.get('tab') || (roleId === 4 ? 'global_analytics' : roleId === 1 ? 'academy' : roleId === 2 ? 'courses' : 'learning');
  const sidebarRef = useRef<HTMLDivElement>(null);

  const router = useRouter();

  useGSAP(() => {
    if (sidebarRef.current) {
      const items = sidebarRef.current.querySelectorAll('li');
      gsap.fromTo(
        items,
        { x: -20, opacity: 0 },
        { x: 0, opacity: 1, duration: 0.5, stagger: 0.05, ease: 'power2.out' }
      );
    }
  }, { scope: sidebarRef });

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <aside ref={sidebarRef} className="w-[260px] hidden lg:flex flex-col border-r-2 border-black bg-white px-6 py-8">
      <nav className="space-y-8 flex-1">
        <NavLinks currentTab={currentTab} />
      </nav>
      <div className="mt-8 border-t-2 border-black pt-6">
        <button onClick={handleLogout} className="flex items-center justify-center gap-2 text-white bg-red-600 px-3 py-3 font-semibold uppercase tracking-widest text-[12px] hover:bg-red-700 transition-colors w-full border-2 border-black">
          <LogOut size={16} />
          Logout
        </button>
      </div>
    </aside>
  );
}

function MobileDashboardNav() {
  const [isOpen, setIsOpen] = useState(false);
  const searchParams = useSearchParams();
  const { roleId, logout } = useAuthStore();
  const router = useRouter();
  const currentTab = searchParams?.get('tab') || (roleId === 4 ? 'global_analytics' : roleId === 1 ? 'academy' : roleId === 2 ? 'courses' : 'learning');

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <div className="lg:hidden border-b-2 border-black bg-white relative z-50">
      <div className="flex items-center justify-between p-4">
        <div className="font-display text-[24px] font-bold uppercase tracking-widest">
          Learnyst
        </div>
        <div className="flex items-center gap-4">
          <button onClick={handleLogout} className="p-2 border-2 border-black text-white bg-red-600 hover:bg-red-700 transition-colors" title="Logout">
            <LogOut size={20} />
          </button>
          <button onClick={() => setIsOpen(!isOpen)} className="p-2 border-2 border-black hover:bg-gray-100 transition-colors">
            {isOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
        </div>
      </div>
      
      {isOpen && (
        <div className="absolute top-full left-0 right-0 bg-white border-b-2 border-black p-6 ">
           <nav className="space-y-8">
             <NavLinks currentTab={currentTab} onNavigate={() => setIsOpen(false)} />
           </nav>
        </div>
      )}
    </div>
  );
}

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { roleId } = useAuthStore();
  return (
    <div className="min-h-screen bg-white bg-graph-paper-light flex flex-col text-black font-ui">
      
      {/* Mobile Navbar specific to the dashboard (hidden on large screens) */}
      <Suspense fallback={<div className="h-16 bg-white border-b-2 border-black lg:hidden" />}>
        <MobileDashboardNav />
      </Suspense>

      <div className="flex-1 flex w-full bg-white relative border-x-0 border-black md:border-x-2">
        <Suspense fallback={<aside className="w-[260px] hidden lg:block border-r-2 border-black px-6 py-8" />}>
          <Sidebar />
        </Suspense>

        
        {/* Main Content */}
        <main className="flex-1 p-4 md:py-8 md:px-6 overflow-x-hidden">
          {children}
        </main>
      </div>
    </div>
  );
}

