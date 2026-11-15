export interface User {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
  role_id: number;
  location: string;
  created_at: string;
}
export interface Course {
  id: number;
  tenant_id: number;
  title: string;
  description: string;
  category: string;
  level: string;
  thumbnail_url: string;
  price: number;
  is_published: boolean;
  instructor: string;
  lesson_count: number;
  student_count: number;
  duration_seconds: number;
  completed_lessons: number;
  progress: number;
  enrolled: boolean;
  updated_at: string;
}
export interface Lesson {
  id: number;
  title: string;
  description: string;
  video_url: string;
  note_url: string;
  duration_seconds: number;
  position: number;
}
export interface Progress {
  video_id: number;
  watch_time_seconds: number;
  completion_percent: number;
}
export interface CourseDetail {
  course: Course;
  videos: Lesson[];
  progress: Progress[];
  can_manage: boolean;
}
export interface Submission {
  id: number;
  student_name: string;
  text: string;
  submitted_at: string;
  marks: number | null;
  feedback: string;
  status: string;
}
export interface Assignment {
  id: number;
  course_id: number;
  course_title: string;
  title: string;
  description: string;
  max_marks: number;
  due_date: string | null;
  submissions: Submission[];
}
export interface Student {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
  course_count: number;
  joined_at: string;
}
