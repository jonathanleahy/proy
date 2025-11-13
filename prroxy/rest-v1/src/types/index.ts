// User types
export interface User {
  id: number;
  name: string;
  username: string;
  email: string;
  address?: {
    street: string;
    suite: string;
    city: string;
    zipcode: string;
    geo: {
      lat: string;
      lng: string;
    };
  };
  phone: string;
  website: string;
  company?: {
    name: string;
    catchPhrase: string;
    bs: string;
  };
}

export interface Post {
  userId: number;
  id: number;
  title: string;
  body: string;
}

export interface Todo {
  userId: number;
  id: number;
  title: string;
  completed: boolean;
}

// Endpoint 1 response
export interface UserResponse {
  id: number;
  name: string;
  username: string;
  email: string;
  phone: string;
  website: string;
}

// Endpoint 2 response
export interface UserSummaryResponse {
  userId: number;
  userName: string;
  email: string;
  postCount: number;
  recentPosts: string[];
  summary: string;
}

// Endpoint 3 request
export interface ReportRequest {
  includeCompleted?: boolean;
  maxPosts?: number;
}

// Endpoint 3 response
export interface UserReportResponse {
  userId: number;
  userName: string;
  email: string;
  stats: {
    totalPosts: number;
    totalTodos: number;
    completedTodos: number;
    pendingTodos: number;
    completionRate: string;
  };
  posts: Array<{
    id: number;
    title: string;
    preview: string;
  }>;
  todos: {
    pending: string[];
    completed: string[];
  };
  generatedAt: string;
}

// Person lookup types
export interface Person {
  firstname: string;
  surname: string;
  dob: string;
  country: string;
}

export interface PersonLookupRequest {
  surname: string;
  dob: string;
}
