import axios from 'axios';
import { UserService } from '../src/services/userService';
import { User, Post, Todo } from '../src/types';

// Mock axios
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('UserService', () => {
  let userService: UserService;

  beforeEach(() => {
    userService = new UserService();
    jest.clearAllMocks();
  });

  describe('getUser', () => {
    it('should fetch and return simplified user data', async () => {
      // Arrange
      const mockUser: User = {
        id: 1,
        name: 'Leanne Graham',
        username: 'Bret',
        email: 'Sincere@april.biz',
        phone: '1-770-736-8031 x56442',
        website: 'hildegard.org',
        address: {
          street: 'Kulas Light',
          suite: 'Apt. 556',
          city: 'Gwenborough',
          zipcode: '92998-3874',
          geo: { lat: '-37.3159', lng: '81.1496' },
        },
        company: {
          name: 'Romaguera-Crona',
          catchPhrase: 'Multi-layered client-server neural-net',
          bs: 'harness real-time e-markets',
        },
      };

      mockedAxios.get.mockResolvedValue({ data: mockUser });

      // Act
      const result = await userService.getUser(1);

      // Assert
      expect(mockedAxios.get).toHaveBeenCalledWith(
        'https://jsonplaceholder.typicode.com/users/1'
      );
      expect(result).toEqual({
        id: 1,
        name: 'Leanne Graham',
        username: 'Bret',
        email: 'Sincere@april.biz',
        phone: '1-770-736-8031 x56442',
        website: 'hildegard.org',
      });
    });

    it('should throw error if user not found', async () => {
      // Arrange
      mockedAxios.get.mockRejectedValue({ response: { status: 404 } });

      // Act & Assert
      await expect(userService.getUser(999)).rejects.toThrow('User not found');
    });

    it('should throw error on API failure', async () => {
      // Arrange
      mockedAxios.get.mockRejectedValue(new Error('Network error'));

      // Act & Assert
      await expect(userService.getUser(1)).rejects.toThrow('Failed to fetch user');
    });
  });

  describe('getUserSummary', () => {
    it('should fetch user and posts, return summary', async () => {
      // Arrange
      const mockUser: User = {
        id: 1,
        name: 'Leanne Graham',
        username: 'Bret',
        email: 'Sincere@april.biz',
        phone: '1-770-736-8031 x56442',
        website: 'hildegard.org',
      };

      const mockPosts: Post[] = [
        {
          userId: 1,
          id: 1,
          title: 'Post 1',
          body: 'Body 1',
        },
        {
          userId: 1,
          id: 2,
          title: 'Post 2',
          body: 'Body 2',
        },
        {
          userId: 1,
          id: 3,
          title: 'Post 3',
          body: 'Body 3',
        },
      ];

      mockedAxios.get
        .mockResolvedValueOnce({ data: mockUser })
        .mockResolvedValueOnce({ data: mockPosts });

      // Act
      const result = await userService.getUserSummary(1);

      // Assert
      expect(mockedAxios.get).toHaveBeenCalledTimes(2);
      expect(mockedAxios.get).toHaveBeenCalledWith(
        'https://jsonplaceholder.typicode.com/users/1'
      );
      expect(mockedAxios.get).toHaveBeenCalledWith(
        'https://jsonplaceholder.typicode.com/posts?userId=1'
      );
      expect(result).toEqual({
        userId: 1,
        userName: 'Leanne Graham',
        email: 'Sincere@april.biz',
        postCount: 3,
        recentPosts: ['Post 1', 'Post 2', 'Post 3'],
        summary: 'User Leanne Graham has written 3 posts',
      });
    });

    it('should handle user with no posts', async () => {
      // Arrange
      const mockUser: User = {
        id: 1,
        name: 'Leanne Graham',
        username: 'Bret',
        email: 'Sincere@april.biz',
        phone: '1-770-736-8031 x56442',
        website: 'hildegard.org',
      };

      mockedAxios.get
        .mockResolvedValueOnce({ data: mockUser })
        .mockResolvedValueOnce({ data: [] });

      // Act
      const result = await userService.getUserSummary(1);

      // Assert
      expect(result.postCount).toBe(0);
      expect(result.recentPosts).toEqual([]);
      expect(result.summary).toBe('User Leanne Graham has written 0 posts');
    });
  });

  describe('getUserReport', () => {
    it('should fetch user, posts, and todos in parallel and return comprehensive report', async () => {
      // Arrange
      const mockUser: User = {
        id: 1,
        name: 'Leanne Graham',
        username: 'Bret',
        email: 'Sincere@april.biz',
        phone: '1-770-736-8031 x56442',
        website: 'hildegard.org',
      };

      const mockPosts: Post[] = [
        { userId: 1, id: 1, title: 'Post 1', body: 'Body 1 with content' },
        { userId: 1, id: 2, title: 'Post 2', body: 'Body 2 with content' },
        { userId: 1, id: 3, title: 'Post 3', body: 'Body 3 with content' },
      ];

      const mockTodos: Todo[] = [
        { userId: 1, id: 1, title: 'Todo 1', completed: false },
        { userId: 1, id: 2, title: 'Todo 2', completed: false },
        { userId: 1, id: 3, title: 'Todo 3', completed: true },
        { userId: 1, id: 4, title: 'Todo 4', completed: false },
        { userId: 1, id: 5, title: 'Todo 5', completed: true },
      ];

      // Mock Promise.all responses
      mockedAxios.get
        .mockResolvedValueOnce({ data: mockUser })
        .mockResolvedValueOnce({ data: mockPosts })
        .mockResolvedValueOnce({ data: mockTodos });

      // Act
      const result = await userService.getUserReport(1, {
        includeCompleted: true,
        maxPosts: 2,
      });

      // Assert
      expect(mockedAxios.get).toHaveBeenCalledTimes(3);
      expect(result.userId).toBe(1);
      expect(result.userName).toBe('Leanne Graham');
      expect(result.stats).toEqual({
        totalPosts: 3,
        totalTodos: 5,
        completedTodos: 2,
        pendingTodos: 3,
        completionRate: '40.0%',
      });
      expect(result.posts).toHaveLength(2);
      expect(result.posts[0]).toEqual({
        id: 1,
        title: 'Post 1',
        preview: 'Body 1 with content',
      });
      expect(result.todos.pending).toEqual(['Todo 1', 'Todo 2', 'Todo 4']);
      expect(result.todos.completed).toEqual(['Todo 3', 'Todo 5']);
      expect(result.generatedAt).toBeDefined();
    });

    it('should limit posts based on maxPosts parameter', async () => {
      // Arrange
      const mockUser: User = {
        id: 1,
        name: 'Leanne Graham',
        username: 'Bret',
        email: 'Sincere@april.biz',
        phone: '1-770-736-8031 x56442',
        website: 'hildegard.org',
      };

      const mockPosts: Post[] = [
        { userId: 1, id: 1, title: 'Post 1', body: 'Body 1' },
        { userId: 1, id: 2, title: 'Post 2', body: 'Body 2' },
        { userId: 1, id: 3, title: 'Post 3', body: 'Body 3' },
      ];

      const mockTodos: Todo[] = [];

      mockedAxios.get
        .mockResolvedValueOnce({ data: mockUser })
        .mockResolvedValueOnce({ data: mockPosts })
        .mockResolvedValueOnce({ data: mockTodos });

      // Act
      const result = await userService.getUserReport(1, { maxPosts: 1 });

      // Assert
      expect(result.posts).toHaveLength(1);
    });

    it('should exclude completed todos when includeCompleted is false', async () => {
      // Arrange
      const mockUser: User = {
        id: 1,
        name: 'Leanne Graham',
        username: 'Bret',
        email: 'Sincere@april.biz',
        phone: '1-770-736-8031 x56442',
        website: 'hildegard.org',
      };

      const mockPosts: Post[] = [];

      const mockTodos: Todo[] = [
        { userId: 1, id: 1, title: 'Todo 1', completed: false },
        { userId: 1, id: 2, title: 'Todo 2', completed: true },
      ];

      mockedAxios.get
        .mockResolvedValueOnce({ data: mockUser })
        .mockResolvedValueOnce({ data: mockPosts })
        .mockResolvedValueOnce({ data: mockTodos });

      // Act
      const result = await userService.getUserReport(1, { includeCompleted: false });

      // Assert
      expect(result.todos.completed).toEqual([]);
      expect(result.todos.pending).toEqual(['Todo 1']);
    });

    it('should handle parallel API call failure', async () => {
      // Arrange
      mockedAxios.get.mockRejectedValue(new Error('Network error'));

      // Act & Assert
      await expect(
        userService.getUserReport(1, {})
      ).rejects.toThrow('Failed to generate user report');
    });
  });
});
