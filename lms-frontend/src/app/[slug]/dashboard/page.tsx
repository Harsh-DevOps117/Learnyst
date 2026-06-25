'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { coreApi } from '@/api/client';
import { useAuthStore } from '@/store/authStore';
import { BookOpen, PlayCircle, Clock } from 'lucide-react';
import Link from 'next/link';

export default function TenantDashboard() {
  const { slug } = useParams();
  const router = useRouter();
  const { userId, token, logout } = useAuthStore();
  
  const [tenant, setTenant] = useState<any>(null);
  const [enrolledCourses, setEnrolledCourses] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token || !userId) {
      router.push(`/${slug}/login`);
      return;
    }

    const fetchData = async () => {
      try {
        // Fetch tenant details first
        const tenantRes = await coreApi.get(`/tenants/slug/${slug}`);
        const tenantData = tenantRes.data.tenant;
        setTenant(tenantData);

        // Fetch enrolled courses for this student and this tenant
        const coursesRes = await coreApi.get(`/students/${userId}/courses?tenant_id=${tenantData.id}`);
        setEnrolledCourses(coursesRes.data || []);
      } catch (err) {
        console.error("Failed to fetch dashboard data", err);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [slug, userId, token, router]);

  if (loading) {
    return (
      <div className="min-h-screen bg-[#fafafa] flex items-center justify-center">
        <div className="w-8 h-8 border-4 border-black border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#fafafa] font-ui text-[#111111]">
      {/* Navbar */}
      <nav className="bg-white border-b border-gray-200 h-16 flex items-center justify-between px-6 md:px-12 sticky top-0 z-50">
        <div className="flex items-center gap-4 cursor-pointer" onClick={() => router.push(`/${slug}`)}>
          {tenant?.logo_url ? (
            <img src={tenant.logo_url} alt="Logo" className="h-8 w-auto object-contain" />
          ) : (
            <span className="font-display text-[18px] font-bold tracking-tight text-gray-900">
              {tenant?.academy_name || 'Academy'}
            </span>
          )}
        </div>
        
        <div className="flex items-center gap-6">
          <Link href="/dashboard" className="text-sm font-medium text-gray-600 hover:text-black">
            Go to Quizzes (Main Site)
          </Link>
          <button 
            onClick={() => {
              logout();
              router.push(`/${slug}/login`);
            }} 
            className="text-sm font-bold text-red-600 hover:text-red-700"
          >
            Sign Out
          </button>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-[1200px] mx-auto px-6 py-12">
        <h1 className="text-3xl font-display font-bold mb-2">My Learning</h1>
        <p className="text-gray-500 mb-10">Pick up right where you left off.</p>

        {enrolledCourses.length === 0 ? (
          <div className="bg-white rounded-2xl border border-gray-200 p-12 text-center shadow-sm">
            <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <BookOpen className="text-gray-400" size={32} />
            </div>
            <h3 className="text-xl font-bold mb-2">No courses yet</h3>
            <p className="text-gray-500 mb-6">You aren't enrolled in any courses in this academy.</p>
            <button 
              onClick={() => router.push(`/${slug}#courses`)}
              className="bg-black text-white px-6 py-2 rounded-full text-sm font-bold hover:bg-gray-800 transition-colors"
            >
              Browse Catalog
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {enrolledCourses.map((course) => (
              <div 
                key={course.id} 
                className="bg-white rounded-2xl border border-gray-200 overflow-hidden shadow-sm hover:shadow-md transition-all group flex flex-col"
              >
                <div className="aspect-video bg-gray-100 relative overflow-hidden">
                  {course.thumbnail_url ? (
                    <img 
                      src={course.thumbnail_url} 
                      alt={course.title} 
                      className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" 
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center bg-gray-200 text-gray-400">
                      <BookOpen size={40} />
                    </div>
                  )}
                  <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                    <PlayCircle size={48} className="text-white" />
                  </div>
                </div>
                
                <div className="p-6 flex flex-col flex-grow">
                  <div className="flex items-center gap-2 text-[12px] font-bold text-gray-500 uppercase tracking-wider mb-2">
                    <span>Course</span>
                  </div>
                  <h3 className="text-lg font-bold text-gray-900 mb-2 leading-tight line-clamp-2">
                    {course.title}
                  </h3>
                  <p className="text-sm text-gray-500 line-clamp-2 mb-6 flex-grow">
                    {course.description}
                  </p>
                  
                  <div className="mt-auto">
                    <button 
                      onClick={() => router.push(`/${slug}/courses/${course.id}`)}
                      className="w-full bg-gray-50 hover:bg-gray-100 text-black border border-gray-200 text-sm font-bold py-2.5 rounded-lg transition-colors"
                    >
                      Continue Learning
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
