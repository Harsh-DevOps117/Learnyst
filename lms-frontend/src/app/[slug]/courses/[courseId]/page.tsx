'use client';

import { useEffect, useState, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { coreApi } from '@/api/client';
import { useAuthStore } from '@/store/authStore';
import { PlayCircle, FileText, UploadCloud, ChevronLeft, Lock } from 'lucide-react';
import Link from 'next/link';
import Hls from 'hls.js';

function HlsPlayer({ src, ...props }: { src: string } & React.VideoHTMLAttributes<HTMLVideoElement>) {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video || !src) return;

    let hls: Hls | null = null;

    if (src.endsWith('.m3u8')) {
      if (Hls.isSupported()) {
        hls = new Hls({
          capLevelToPlayerSize: true,
          autoStartLoad: true
        });
        hls.loadSource(src);
        hls.attachMedia(video);
      } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
        // Native HLS support (Safari)
        video.src = src;
      }
    } else {
      // Standard MP4 fallback
      video.src = src;
    }

    return () => {
      if (hls) {
        hls.destroy();
      }
    };
  }, [src]);

  return <video ref={videoRef} {...props} />;
}

export default function CoursePlayerPage() {
  const { slug, courseId } = useParams();
  const router = useRouter();
  const { userId, token } = useAuthStore();
  
  const [tenant, setTenant] = useState<any>(null);
  const [courseContent, setCourseContent] = useState<{videos: any[], assignments: any[], notes: any[]}>({ videos: [], assignments: [], notes: [] });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  const [activeVideo, setActiveVideo] = useState<any>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'notes' | 'assignments'>('overview');

  useEffect(() => {
    if (!token || !userId) {
      router.push(`/${slug}/login`);
      return;
    }

    const fetchData = async () => {
      try {
        const tenantRes = await coreApi.get(`/tenants/slug/${slug}`);
        setTenant(tenantRes.data.tenant);

        // Fetch course content (videos, assignments, notes)
        const contentRes = await coreApi.get(`/courses/${courseId}/consume?student_id=${userId}`);
        const content = contentRes.data;
        setCourseContent(content);
        
        if (content.videos && content.videos.length > 0) {
          setActiveVideo(content.videos[0]);
        }
      } catch (err: any) {
        if (err.response?.status === 403) {
          setError("Access Denied. You are not enrolled in this course.");
        } else {
          setError("Failed to load course content.");
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [slug, courseId, userId, token, router]);

  if (loading) {
    return (
      <div className="min-h-screen bg-[#fafafa] flex items-center justify-center">
        <div className="w-8 h-8 border-4 border-black border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-[#fafafa] flex flex-col items-center justify-center p-6 font-ui">
        <div className="w-16 h-16 bg-red-100 text-red-500 flex items-center justify-center rounded-full mb-4">
          <Lock size={32} />
        </div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">{error}</h1>
        <button 
          onClick={() => router.push(`/${slug}/dashboard`)}
          className="mt-6 px-6 py-2 bg-black text-white rounded-full text-sm font-bold hover:bg-gray-800"
        >
          Return to Dashboard
        </button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#fafafa] font-ui text-[#111111] flex flex-col">
      {/* Navbar */}
      <nav className="bg-white border-b border-gray-200 h-16 flex items-center px-4 shrink-0">
        <button 
          onClick={() => router.push(`/${slug}/dashboard`)}
          className="flex items-center gap-2 text-sm font-medium text-gray-600 hover:text-black mr-6"
        >
          <ChevronLeft size={16} /> Back
        </button>
        <div className="h-6 w-px bg-gray-200 mr-6"></div>
        <div className="font-display font-bold truncate">Course Player</div>
      </nav>

      <div className="flex flex-1 overflow-hidden">
        {/* Main Content Area */}
        <div className="flex-1 flex flex-col overflow-y-auto">
          {/* Video Player */}
          <div className="w-full bg-black aspect-video flex items-center justify-center">
            {activeVideo ? (
              <HlsPlayer
                key={activeVideo.id} 
                src={activeVideo.video_url} 
                controls 
                controlsList="nodownload"
                className="w-full h-full object-contain bg-black"
                autoPlay
              />
            ) : (
              <div className="text-gray-500 font-medium">No video selected or available</div>
            )}
          </div>

          {/* Content Below Video */}
          <div className="max-w-[1000px] w-full mx-auto p-6 md:p-10">
            <h1 className="text-2xl font-bold mb-6">{activeVideo?.title || 'Course Overview'}</h1>
            
            {/* Tabs */}
            <div className="flex gap-8 border-b border-gray-200 mb-8">
              <button 
                onClick={() => setActiveTab('overview')}
                className={`pb-4 text-sm font-bold transition-colors relative ${activeTab === 'overview' ? 'text-black' : 'text-gray-400 hover:text-gray-600'}`}
              >
                Overview
                {activeTab === 'overview' && <div className="absolute bottom-0 left-0 w-full h-0.5 bg-black rounded-t-full"></div>}
              </button>
              <button 
                onClick={() => setActiveTab('notes')}
                className={`pb-4 text-sm font-bold transition-colors relative ${activeTab === 'notes' ? 'text-black' : 'text-gray-400 hover:text-gray-600'}`}
              >
                Notes
                {activeTab === 'notes' && <div className="absolute bottom-0 left-0 w-full h-0.5 bg-black rounded-t-full"></div>}
              </button>
              <button 
                onClick={() => setActiveTab('assignments')}
                className={`pb-4 text-sm font-bold transition-colors relative ${activeTab === 'assignments' ? 'text-black' : 'text-gray-400 hover:text-gray-600'}`}
              >
                Assignments
                {activeTab === 'assignments' && <div className="absolute bottom-0 left-0 w-full h-0.5 bg-black rounded-t-full"></div>}
              </button>
            </div>

            {/* Tab Content */}
            <div className="min-h-[300px]">
              {activeTab === 'overview' && (
                <div className="prose prose-sm max-w-none text-gray-600">
                  <p>{activeVideo?.description || 'No description available for this video.'}</p>
                </div>
              )}

              {activeTab === 'notes' && (
                <div className="flex flex-col gap-4">
                  {courseContent.notes.length === 0 ? (
                    <p className="text-gray-500 text-sm">No notes uploaded for this course yet.</p>
                  ) : (
                    courseContent.notes.map((note) => (
                      <div key={note.ID} className="flex items-center justify-between p-4 bg-white border border-gray-200 rounded-xl hover:shadow-sm transition-shadow">
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 bg-blue-50 text-blue-600 rounded-lg flex items-center justify-center">
                            <FileText size={20} />
                          </div>
                          <span className="font-medium text-sm text-gray-900">{note.Title}</span>
                        </div>
                        <a 
                          href={note.FileUrl} 
                          target="_blank" 
                          rel="noopener noreferrer"
                          className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-900 text-sm font-bold rounded-lg transition-colors"
                        >
                          Download
                        </a>
                      </div>
                    ))
                  )}
                </div>
              )}

              {activeTab === 'assignments' && (
                <div className="flex flex-col gap-4">
                  {courseContent.assignments.length === 0 ? (
                    <p className="text-gray-500 text-sm">No assignments posted for this course yet.</p>
                  ) : (
                    courseContent.assignments.map((assignment) => (
                      <div key={assignment.ID} className="p-6 bg-white border border-gray-200 rounded-xl hover:shadow-sm transition-shadow">
                        <div className="flex items-center gap-3 mb-4">
                          <div className="w-10 h-10 bg-purple-50 text-purple-600 rounded-lg flex items-center justify-center">
                            <UploadCloud size={20} />
                          </div>
                          <div>
                            <h3 className="font-bold text-gray-900">{assignment.Title}</h3>
                            <p className="text-[12px] font-medium text-gray-500">Max Marks: {assignment.MaxMarks}</p>
                          </div>
                        </div>
                        <p className="text-sm text-gray-600 mb-6">{assignment.Description}</p>
                        <div className="flex gap-3">
                          {assignment.AttachmentUrl && (
                            <a 
                              href={assignment.AttachmentUrl} 
                              target="_blank" 
                              rel="noopener noreferrer"
                              className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-900 text-sm font-bold rounded-lg transition-colors"
                            >
                              Download Brief
                            </a>
                          )}
                          {/* We leave submission logic for a future sprint */}
                          <button className="px-4 py-2 bg-black hover:bg-gray-800 text-white text-sm font-bold rounded-lg transition-colors">
                            Submit Work
                          </button>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="w-[350px] bg-white border-l border-gray-200 flex flex-col shrink-0">
          <div className="p-4 border-b border-gray-200">
            <h2 className="font-bold text-lg">Course Content</h2>
            <p className="text-sm text-gray-500">{courseContent.videos.length} Lectures</p>
          </div>
          
          <div className="flex-1 overflow-y-auto">
            {courseContent.videos.map((video, index) => (
              <button
                key={video.id}
                onClick={() => setActiveVideo(video)}
                className={`w-full text-left p-4 flex gap-3 border-b border-gray-100 transition-colors ${activeVideo?.id === video.id ? 'bg-gray-50' : 'hover:bg-gray-50'}`}
              >
                <div className="mt-1">
                  <PlayCircle size={18} className={activeVideo?.id === video.id ? 'text-black' : 'text-gray-400'} />
                </div>
                <div>
                  <h4 className={`text-sm leading-tight mb-1 ${activeVideo?.id === video.id ? 'font-bold text-black' : 'font-medium text-gray-700'}`}>
                    {index + 1}. {video.title}
                  </h4>
                  <span className="text-[12px] text-gray-500 font-medium">Video Lecture</span>
                </div>
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
