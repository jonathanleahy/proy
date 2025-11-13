import axios from 'axios';
import {
  User,
  Post,
  Todo,
  UserResponse,
  UserSummaryResponse,
  ReportRequest,
  UserReportResponse,
} from '../types';

// Use proxy if PROXY_URL is set, otherwise use direct URL
const PROXY_URL = process.env.PROXY_URL;
const BASE_URL = PROXY_URL
  ? `${PROXY_URL}?target=jsonplaceholder.typicode.com`
  : 'https://jsonplaceholder.typicode.com';

// Configure axios to disable compression when using proxy
// This ensures recordings are stored uncompressed for compatibility
const axiosConfig = PROXY_URL
  ? {
      decompress: false,
      headers: {
        'Accept-Encoding': 'identity',
      },
    }
  : {};

export class UserService {
  /**
   * Endpoint 1: Simple - Get user data
   */
  async getUser(userId: number): Promise<UserResponse> {
    try {
      const response = await axios.get<User>(`${BASE_URL}/users/${userId}`, axiosConfig);
      const user = response.data;

      // Return simplified user data
      return {
        id: user.id,
        name: user.name,
        username: user.username,
        email: user.email,
        phone: user.phone,
        website: user.website,
      };
    } catch (error: any) {
      if (error.response?.status === 404) {
        throw new Error('User not found');
      }
      throw new Error('Failed to fetch user');
    }
  }

  /**
   * Endpoint 2: Medium complexity - Get user summary with posts
   */
  async getUserSummary(userId: number): Promise<UserSummaryResponse> {
    try {
      // Call 1: Fetch user
      const userResponse = await axios.get<User>(`${BASE_URL}/users/${userId}`, axiosConfig);
      const user = userResponse.data;

      // Call 2: Fetch user's posts
      const postsResponse = await axios.get<Post[]>(
        `${BASE_URL}/posts?userId=${userId}`,
        axiosConfig
      );
      const posts = postsResponse.data;

      // Manipulate data
      const postTitles = posts.map((post) => post.title);

      return {
        userId: user.id,
        userName: user.name,
        email: user.email,
        postCount: posts.length,
        recentPosts: postTitles,
        summary: `User ${user.name} has written ${posts.length} posts`,
      };
    } catch (error) {
      throw new Error('Failed to fetch user summary');
    }
  }

  /**
   * Endpoint 3: Complex - Get comprehensive user report with parallel calls
   */
  async getUserReport(
    userId: number,
    options: ReportRequest
  ): Promise<UserReportResponse> {
    try {
      // Parallel calls to 3 different endpoints
      const [userResponse, postsResponse, todosResponse] = await Promise.all([
        axios.get<User>(`${BASE_URL}/users/${userId}`, axiosConfig),
        axios.get<Post[]>(`${BASE_URL}/posts?userId=${userId}`, axiosConfig),
        axios.get<Todo[]>(`${BASE_URL}/todos?userId=${userId}`, axiosConfig),
      ]);

      const user = userResponse.data;
      const posts = postsResponse.data;
      const todos = todosResponse.data;

      // Complex data manipulation
      const { includeCompleted = true, maxPosts } = options;

      // Calculate todo statistics
      const completedTodos = todos.filter((todo) => todo.completed);
      const pendingTodos = todos.filter((todo) => !todo.completed);
      const completionRate = todos.length > 0
        ? ((completedTodos.length / todos.length) * 100).toFixed(1)
        : '0.0';

      // Limit posts if maxPosts specified
      const limitedPosts = maxPosts ? posts.slice(0, maxPosts) : posts;

      // Format posts with preview
      const formattedPosts = limitedPosts.map((post) => ({
        id: post.id,
        title: post.title,
        preview: post.body,
      }));

      // Format todos based on includeCompleted option
      const todosResult = {
        pending: pendingTodos.map((todo) => todo.title),
        completed: includeCompleted
          ? completedTodos.map((todo) => todo.title)
          : [],
      };

      return {
        userId: user.id,
        userName: user.name,
        email: user.email,
        stats: {
          totalPosts: posts.length,
          totalTodos: todos.length,
          completedTodos: completedTodos.length,
          pendingTodos: pendingTodos.length,
          completionRate: `${completionRate}%`,
        },
        posts: formattedPosts,
        todos: todosResult,
        generatedAt: new Date().toISOString(),
      };
    } catch (error) {
      throw new Error('Failed to generate user report');
    }
  }
}
