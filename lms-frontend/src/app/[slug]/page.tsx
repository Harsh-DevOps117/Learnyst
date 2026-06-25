"use client";

import { useEffect, useState, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import { coreApi } from "@/api/client";
import gsap from "gsap";
import { useGSAP } from "@gsap/react";
import { ArrowRight, Star, Users, PlayCircle } from "lucide-react";

gsap.registerPlugin(useGSAP);

export default function AcademyPublicPage() {
  const { slug } = useParams();
  const router = useRouter();
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const containerRef = useRef<HTMLDivElement>(null);
  const navRef = useRef<HTMLElement>(null);

  useEffect(() => {
    coreApi
      .get(`/tenants/slug/${slug}`)
      .then((res) => {
        setData(res.data);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.response?.data?.error || "Failed to load academy");
        setLoading(false);
      });
  }, [slug]);

  useGSAP(() => {
    if (!loading && data) {
      // Elegant staggered reveal for all major sections
      gsap.from(".reveal-up", {
        y: 60,
        opacity: 0,
        duration: 1.2,
        stagger: 0.2,
        ease: "power4.out",
        clearProps: "all"
      });

      // Floating animation for decorative blobs
      gsap.to(".float-blob", {
        y: -30,
        x: 20,
        rotation: 5,
        duration: 4,
        yoyo: true,
        repeat: -1,
        ease: "sine.inOut",
        stagger: 0.5
      });
    }
  }, { scope: containerRef, dependencies: [loading, data] });

  // Add scroll event listener for navbar shadow
  useEffect(() => {
    const handleScroll = () => {
      if (window.scrollY > 50) {
        navRef.current?.classList.add("shadow-sm", "bg-white/90", "backdrop-blur-xl");
        navRef.current?.classList.remove("bg-transparent", "backdrop-blur-none");
      } else {
        navRef.current?.classList.remove("shadow-sm", "bg-white/90", "backdrop-blur-xl");
        navRef.current?.classList.add("bg-transparent", "backdrop-blur-none");
      }
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  if (loading) return (
    <div className="min-h-screen flex items-center justify-center bg-[#fafafa]">
      <div className="w-8 h-8 border-4 border-black border-t-transparent rounded-full animate-spin"></div>
    </div>
  );
  
  if (error || !data) return (
    <div className="min-h-screen flex items-center justify-center bg-[#fafafa]">
      <h1 className="text-2xl font-display font-bold">404 - Academy Not Found</h1>
    </div>
  );

  const { tenant, courses } = data;
  const owner = tenant.owner;

  // Use mock courses if the tenant hasn't uploaded any yet so the grid still looks good!
  const displayCourses = courses && courses.length > 0 ? courses : [
    { id: 'demo1', title: 'The Complete Masterclass', description: 'Learn the fundamentals and advanced techniques in this comprehensive guide.', thumbnail_url: '' },
    { id: 'demo2', title: 'Advanced Strategies', description: 'Take your skills to the next level with industry-proven strategies.', thumbnail_url: '' },
    { id: 'demo3', title: 'Real-World Projects', description: 'Apply your knowledge by building 5 real-world projects from scratch.', thumbnail_url: '' }
  ];

  return (
    <div ref={containerRef} className="min-h-screen bg-[#fafafa] text-[#111111] font-ui antialiased overflow-hidden">
      {/* Decorative Background Blobs */}
      <div className="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none z-0">
        <div className="float-blob absolute top-[-10%] left-[-10%] w-[500px] h-[500px] rounded-full bg-blue-100/50 blur-[100px]" />
        <div className="float-blob absolute top-[20%] right-[-5%] w-[400px] h-[400px] rounded-full bg-orange-100/50 blur-[100px]" />
        <div className="float-blob absolute bottom-[-10%] left-[20%] w-[600px] h-[600px] rounded-full bg-green-100/40 blur-[120px]" />
      </div>

      {/* Premium Sticky Navigation */}
      <nav ref={navRef} className="fixed top-0 left-0 right-0 h-20 flex items-center justify-between px-6 md:px-12 z-50 transition-all duration-300 bg-transparent border-b border-gray-200/50">
        <div className="flex items-center gap-4 cursor-pointer hover:opacity-80 transition-opacity">
          {tenant.logo_url && (
            <img src={tenant.logo_url} alt="Logo" className="h-10 w-auto object-contain" />
          )}
          <span className="font-display text-[20px] font-bold tracking-tight text-gray-900">{tenant.academy_name}</span>
        </div>
        
        <div className="hidden md:flex items-center gap-8 bg-white/50 px-6 py-2 rounded-full border border-gray-200 backdrop-blur-md shadow-sm">
          <a href="#courses" className="text-[14px] font-semibold text-gray-600 hover:text-black transition-colors">Courses</a>
          <a href="#instructor" className="text-[14px] font-semibold text-gray-600 hover:text-black transition-colors">Instructor</a>
          <a href="#about" className="text-[14px] font-semibold text-gray-600 hover:text-black transition-colors">About</a>
        </div>

        <div className="flex items-center gap-4">
          <button onClick={() => router.push(`/${slug}/login`)} className="hidden sm:block text-[14px] font-bold text-gray-600 hover:text-black transition-colors px-4 py-2">
            Log In
          </button>
          <button onClick={() => router.push(`/${slug}/signup`)} className="bg-black hover:bg-gray-800 text-white text-[14px] font-bold px-6 py-2.5 rounded-full shadow-lg hover:shadow-xl hover:-translate-y-0.5 active:scale-95 transition-all duration-300 flex items-center gap-2 group">
            Sign Up <ArrowRight size={16} className="group-hover:translate-x-1 transition-transform" />
          </button>
        </div>
      </nav>

      <main className="relative z-10 pt-32 pb-24 px-6 md:px-12 max-w-[1440px] mx-auto">
        
        {/* Spectacular Hero Section */}
        <section className="reveal-up text-center max-w-4xl mx-auto pt-16 pb-24">
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-white border border-gray-200 shadow-sm mb-8 animate-fade-in">
            <span className="flex h-2 w-2 rounded-full bg-green-500"></span>
            <span className="text-[12px] font-bold text-gray-600 uppercase tracking-widest">Accepting New Students</span>
          </div>
          
          <h1 className="font-display text-[56px] md:text-[80px] leading-[1.05] font-extrabold tracking-tighter text-gray-900 mb-8">
            Master your craft at <br className="hidden md:block"/>
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-gray-900 via-gray-700 to-gray-500">prototype speed.</span>
          </h1>
          
          <p className="text-[20px] md:text-[24px] leading-relaxed text-gray-600 font-medium max-w-2xl mx-auto mb-12">
            {tenant.description || `Welcome to ${tenant.academy_name}. Learn directly from industry experts and accelerate your career with our highly specialized content.`}
          </p>
          
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <button onClick={() => router.push(`/${slug}/signup`)} className="w-full sm:w-auto bg-black text-white text-[16px] font-bold px-8 py-4 rounded-full shadow-2xl hover:shadow-[0_20px_40px_-15px_rgba(0,0,0,0.5)] hover:-translate-y-1 active:scale-95 transition-all duration-300 flex items-center justify-center gap-2 group">
              Start Learning Free <ArrowRight size={18} className="group-hover:translate-x-1 transition-transform" />
            </button>
            <button onClick={() => router.push('#courses')} className="w-full sm:w-auto bg-white text-gray-900 text-[16px] font-bold px-8 py-4 rounded-full border border-gray-200 shadow-sm hover:shadow-md hover:-translate-y-1 active:scale-95 transition-all duration-300">
              View Courses
            </button>
          </div>
        </section>

        {/* Feature Banner / Signature Airtable-style Card */}
        {tenant.banner_url && (
          <section className="reveal-up mb-32">
            <div className="relative w-full h-[500px] md:h-[600px] rounded-[32px] overflow-hidden group cursor-pointer shadow-2xl" onClick={() => router.push('/signup')}>
              <div className="absolute inset-0 bg-black/40 z-10 group-hover:bg-black/30 transition-colors duration-500"></div>
              <img src={tenant.banner_url} alt="Academy Banner" className="absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition-transform duration-1000 ease-out" />
              
              <div className="absolute inset-0 z-20 flex flex-col justify-end p-8 md:p-16">
                <div className="max-w-2xl transform group-hover:-translate-y-2 transition-transform duration-500">
                  <h2 className="font-display text-[40px] md:text-[56px] font-bold text-white leading-tight mb-4 tracking-tight drop-shadow-lg">
                    Elevate your potential.
                  </h2>
                  <p className="text-[18px] md:text-[20px] text-white/90 font-medium mb-8 max-w-xl drop-shadow-md">
                    Join an elite community of learners and get access to exclusive resources, community forums, and hands-on projects.
                  </p>
                  <div className="inline-flex items-center gap-2 bg-white/20 backdrop-blur-md text-white px-6 py-3 rounded-full font-bold text-[14px] uppercase tracking-widest border border-white/30 hover:bg-white/30 transition-colors">
                    Join the Academy <ArrowRight size={16} />
                  </div>
                </div>
              </div>
            </div>
          </section>
        )}

        {/* Sleek Minimalist Value Props */}
        <section className="reveal-up mb-32 border-t border-b border-gray-200 py-20">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-16 md:gap-8 divide-y md:divide-y-0 md:divide-x divide-gray-200">
            <div className="px-4 md:px-8">
              <span className="text-[64px] font-display font-light text-gray-200 mb-2 block leading-none">01</span>
              <h3 className="font-display text-[24px] font-bold text-gray-900 mb-4">Structured Learning</h3>
              <p className="text-gray-600 leading-relaxed font-medium">Follow a proven path from beginner to expert with our meticulously crafted course modules and hands-on exercises.</p>
            </div>
            <div className="px-4 md:px-8 pt-12 md:pt-0">
              <span className="text-[64px] font-display font-light text-gray-200 mb-2 block leading-none">02</span>
              <h3 className="font-display text-[24px] font-bold text-gray-900 mb-4">Elite Community</h3>
              <p className="text-gray-600 leading-relaxed font-medium">Connect with ambitious peers, collaborate on projects, and get feedback directly from instructors and alumni.</p>
            </div>
            <div className="px-4 md:px-8 pt-12 md:pt-0">
              <span className="text-[64px] font-display font-light text-gray-200 mb-2 block leading-none">03</span>
              <h3 className="font-display text-[24px] font-bold text-gray-900 mb-4">Industry Recognition</h3>
              <p className="text-gray-600 leading-relaxed font-medium">Earn verifiable certificates upon completion that signal your expertise to top employers and clients worldwide.</p>
            </div>
          </div>
        </section>

        {/* Premium Courses Grid */}
        <section id="courses" className="reveal-up mb-32">
          <div className="flex flex-col md:flex-row md:items-end justify-between mb-12 gap-6">
            <div>
              <h2 className="font-display text-[48px] font-bold text-gray-900 tracking-tight leading-none mb-4">Courses</h2>
              <p className="text-[18px] text-gray-600 font-medium">Transformative content designed for rapid mastery.</p>
            </div>
            <button onClick={() => router.push('/signup')} className="text-[14px] font-bold text-gray-900 border-b-2 border-gray-900 pb-1 hover:text-gray-600 hover:border-gray-600 transition-colors inline-flex items-center gap-2">
              View All <ArrowRight size={16} />
            </button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {displayCourses.map((course: any, idx: number) => (
              <div key={course.id || idx} onClick={() => router.push('/signup')} className="group flex flex-col bg-white rounded-[24px] overflow-hidden border border-gray-200 shadow-sm hover:shadow-2xl hover:-translate-y-2 transition-all duration-500 cursor-pointer">
                <div className="relative aspect-[4/3] bg-gray-100 overflow-hidden">
                  {course.thumbnail_url ? (
                    <img src={course.thumbnail_url} alt={course.title} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-700" />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-gray-400">
                      <PlayCircle size={48} strokeWidth={1} />
                    </div>
                  )}
                  {/* Floating badge */}
                  <div className="absolute top-4 right-4 bg-white/90 backdrop-blur-md px-3 py-1 rounded-full text-[12px] font-bold text-gray-900 shadow-sm flex items-center gap-1">
                    <Star size={12} className="text-orange-400 fill-orange-400" /> Premium
                  </div>
                </div>
                
                <div className="p-8 flex flex-col flex-grow">
                  <h3 className="font-display text-[24px] font-bold text-gray-900 mb-3 leading-tight group-hover:text-blue-600 transition-colors">{course.title}</h3>
                  <p className="text-gray-600 text-[15px] leading-relaxed line-clamp-3 mb-8 flex-grow">{course.description}</p>
                  
                  <div className="flex items-center justify-between pt-6 border-t border-gray-100 mt-auto">
                    <span className="text-[14px] font-bold text-gray-900 uppercase tracking-widest">Enroll Now</span>
                    <div className="w-10 h-10 rounded-full bg-gray-50 flex items-center justify-center group-hover:bg-blue-600 group-hover:text-white transition-colors">
                      <ArrowRight size={18} />
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Editorial Instructor Section */}
        {owner && (
          <section id="instructor" className="reveal-up mb-32 bg-gray-900 rounded-[32px] overflow-hidden relative shadow-2xl">
            <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/cubes.png')] opacity-10 mix-blend-overlay"></div>
            <div className="relative z-10 flex flex-col md:flex-row items-center">
              <div className="w-full md:w-1/2 p-12 md:p-20 flex flex-col justify-center">
                <h2 className="font-display text-[48px] font-bold text-white tracking-tight mb-6 leading-none">
                  Learn directly from the expert.
                </h2>
                <p className="text-[18px] leading-relaxed text-gray-300 font-medium mb-10">
                  Hi, I'm <strong className="text-white">{owner.first_name} {owner.last_name}</strong>. I built {tenant.academy_name} to share the exact strategies, frameworks, and skills I've used throughout my career. Cut through the noise and learn what actually works.
                </p>
                <button onClick={() => router.push('/signup')} className="self-start bg-white text-gray-900 text-[16px] font-bold px-8 py-4 rounded-full hover:bg-gray-100 active:scale-95 transition-all duration-300 shadow-xl">
                  Meet the Instructor
                </button>
              </div>
              <div className="w-full md:w-1/2 h-[400px] md:h-[600px] relative">
                {owner.profile_image_url ? (
                  <img src={owner.profile_image_url} alt={owner.first_name} className="w-full h-full object-cover object-center" />
                ) : (
                  <div className="w-full h-full bg-gray-800 flex items-center justify-center">
                    <Users size={64} className="text-gray-600" />
                  </div>
                )}
                {/* Gradient fade to blend image with dark background */}
                <div className="absolute inset-0 bg-gradient-to-r from-gray-900 to-transparent w-1/3 hidden md:block"></div>
                <div className="absolute inset-0 bg-gradient-to-t from-gray-900 to-transparent h-1/3 mt-auto md:hidden"></div>
              </div>
            </div>
          </section>
        )}

        {/* Minimalist Footer CTA */}
        <section className="reveal-up text-center py-24 border-t border-gray-200">
          <h2 className="font-display text-[40px] font-bold text-gray-900 mb-8 tracking-tight">Ready to level up?</h2>
          <button onClick={() => router.push('/signup')} className="bg-black text-white text-[18px] font-bold px-10 py-5 rounded-full shadow-2xl hover:-translate-y-1 active:scale-95 transition-all duration-300 inline-flex items-center gap-3">
            Create Free Account <ArrowRight size={20} />
          </button>
        </section>
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-200 py-16 px-6">
        <div className="max-w-[1440px] mx-auto flex flex-col md:flex-row justify-between gap-12">
          
          <div className="flex flex-col gap-6 max-w-sm">
            <div className="flex items-center gap-3">
              {tenant.logo_url ? (
                <img src={tenant.logo_url} alt="Logo" className="h-8 w-auto grayscale opacity-80" />
              ) : (
                <span className="font-display text-[24px] font-bold text-gray-900 tracking-tight">{tenant.academy_name}</span>
              )}
            </div>
            <p className="text-gray-500 text-[15px] leading-relaxed">
              {tenant.description || `The premier destination for mastering your craft and accelerating your career.`}
            </p>
          </div>

          <div className="flex flex-col md:flex-row gap-12 md:gap-24">
            <div className="flex flex-col gap-4">
              <h4 className="font-display font-bold text-gray-900 uppercase tracking-widest text-[13px]">Contact Us</h4>
              {owner?.email && (
                <a href={`mailto:${owner.email}`} className="text-gray-600 hover:text-blue-600 transition-colors text-[15px] font-medium">
                  {owner.email}
                </a>
              )}
              {owner?.phone && (
                <a href={`tel:${owner.phone}`} className="text-gray-600 hover:text-blue-600 transition-colors text-[15px] font-medium">
                  {owner.phone}
                </a>
              )}
              {!owner?.email && !owner?.phone && (
                <span className="text-gray-500 text-[15px] font-medium">No contact info available</span>
              )}
            </div>

            <div className="flex flex-col gap-4">
              <h4 className="font-display font-bold text-gray-900 uppercase tracking-widest text-[13px]">Legal</h4>
              <a href="#" className="text-gray-600 hover:text-gray-900 transition-colors text-[15px] font-medium">Terms of Service</a>
              <a href="#" className="text-gray-600 hover:text-gray-900 transition-colors text-[15px] font-medium">Privacy Policy</a>
            </div>
          </div>
          
        </div>
        
        <div className="max-w-[1440px] mx-auto mt-16 pt-8 border-t border-gray-100 flex flex-col md:flex-row items-center justify-between gap-4">
          <span className="text-gray-400 text-[14px] font-medium">© {new Date().getFullYear()} {tenant.academy_name}. All rights reserved.</span>
          <span className="text-gray-400 text-[14px] font-medium">Powered by <strong className="text-gray-600">Learnyst</strong></span>
        </div>
      </footer>
    </div>
  );
}
