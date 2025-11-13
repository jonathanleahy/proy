import request from 'supertest';
import express from 'express';
import { createUserRouter } from '../src/routes/userRoutes';
import { UserService } from '../src/services/userService';

// Mock the UserService
jest.mock('../src/services/userService');

describe('User Routes', () => {
  let app: any;
  let mockUserService: jest.Mocked<UserService>;

  beforeEach(() => {
    // Create mocked UserService instance
    mockUserService = new UserService() as jest.Mocked<UserService>;

    // Create Express app with mocked service
    app = express();
    app.use(express.json());
    app.use('/api', createUserRouter(mockUserService));

    jest.clearAllMocks();
  });

  describe('GET /api/user/:id', () => {
    it('should return user data for valid user id', async () => {
      // Arrange
      const mockUserData = {
        id: 1,
        name: 'Leanne Graham',
        username: 'Bret',
        email: 'Sincere@april.biz',
        phone: '1-770-736-8031 x56442',
        website: 'hildegard.org',
      };

      mockUserService.getUser = jest.fn().mockResolvedValue(mockUserData);

      // Act
      const response = await request(app).get('/api/user/1');

      // Assert
      expect(response.status).toBe(200);
      expect(response.body).toEqual(mockUserData);
    });

    it('should return 404 for non-existent user', async () => {
      // Arrange
      mockUserService.getUser = jest
        .fn()
        .mockRejectedValue(new Error('User not found'));

      // Act
      const response = await request(app).get('/api/user/999');

      // Assert
      expect(response.status).toBe(404);
      expect(response.body).toEqual({ error: 'User not found' });
    });

    it('should return 500 for server errors', async () => {
      // Arrange
      mockUserService.getUser = jest
        .fn()
        .mockRejectedValue(new Error('Failed to fetch user'));

      // Act
      const response = await request(app).get('/api/user/1');

      // Assert
      expect(response.status).toBe(500);
      expect(response.body).toEqual({ error: 'Failed to fetch user' });
    });

    it('should handle invalid user id format', async () => {
      // Act
      const response = await request(app).get('/api/user/invalid');

      // Assert
      expect(response.status).toBe(400);
      expect(response.body).toEqual({ error: 'Invalid user ID' });
    });
  });

  describe('GET /api/user/:id/summary', () => {
    it('should return user summary with posts', async () => {
      // Arrange
      const mockSummary = {
        userId: 1,
        userName: 'Leanne Graham',
        email: 'Sincere@april.biz',
        postCount: 3,
        recentPosts: ['Post 1', 'Post 2', 'Post 3'],
        summary: 'User Leanne Graham has written 3 posts',
      };

      mockUserService.getUserSummary = jest.fn().mockResolvedValue(mockSummary);

      // Act
      const response = await request(app).get('/api/user/1/summary');

      // Assert
      expect(response.status).toBe(200);
      expect(response.body).toEqual(mockSummary);
    });

    it('should return 500 for server errors', async () => {
      // Arrange
      mockUserService.getUserSummary = jest
        .fn()
        .mockRejectedValue(new Error('Failed to fetch user summary'));

      // Act
      const response = await request(app).get('/api/user/999/summary');

      // Assert
      expect(response.status).toBe(500);
      expect(response.body).toEqual({ error: 'Failed to fetch user summary' });
    });

    it('should handle invalid user id format', async () => {
      // Act
      const response = await request(app).get('/api/user/invalid/summary');

      // Assert
      expect(response.status).toBe(400);
      expect(response.body).toEqual({ error: 'Invalid user ID' });
    });
  });

  describe('POST /api/user/:id/report', () => {
    it('should return comprehensive user report with default options', async () => {
      // Arrange
      const mockReport = {
        userId: 1,
        userName: 'Leanne Graham',
        email: 'Sincere@april.biz',
        stats: {
          totalPosts: 3,
          totalTodos: 5,
          completedTodos: 2,
          pendingTodos: 3,
          completionRate: '40.0%',
        },
        posts: [
          { id: 1, title: 'Post 1', preview: 'Body 1' },
          { id: 2, title: 'Post 2', preview: 'Body 2' },
        ],
        todos: {
          pending: ['Todo 1', 'Todo 2'],
          completed: ['Todo 3'],
        },
        generatedAt: '2025-11-01T15:00:00.000Z',
      };

      mockUserService.getUserReport = jest.fn().mockResolvedValue(mockReport);

      // Act
      const response = await request(app).post('/api/user/1/report').send({});

      // Assert
      expect(response.status).toBe(200);
      expect(response.body).toEqual(mockReport);
    });

    it('should accept and pass report options to service', async () => {
      // Arrange
      const reportOptions = {
        includeCompleted: false,
        maxPosts: 2,
      };

      const mockReport = {
        userId: 1,
        userName: 'Leanne Graham',
        email: 'Sincere@april.biz',
        stats: {
          totalPosts: 3,
          totalTodos: 5,
          completedTodos: 2,
          pendingTodos: 3,
          completionRate: '40.0%',
        },
        posts: [
          { id: 1, title: 'Post 1', preview: 'Body 1' },
          { id: 2, title: 'Post 2', preview: 'Body 2' },
        ],
        todos: {
          pending: ['Todo 1', 'Todo 2'],
          completed: [],
        },
        generatedAt: '2025-11-01T15:00:00.000Z',
      };

      mockUserService.getUserReport = jest.fn().mockResolvedValue(mockReport);

      // Act
      const response = await request(app)
        .post('/api/user/1/report')
        .send(reportOptions);

      // Assert
      expect(response.status).toBe(200);
      expect(response.body.todos.completed).toEqual([]);
    });

    it('should return 500 for server errors', async () => {
      // Arrange
      mockUserService.getUserReport = jest
        .fn()
        .mockRejectedValue(new Error('Failed to generate user report'));

      // Act
      const response = await request(app).post('/api/user/1/report').send({});

      // Assert
      expect(response.status).toBe(500);
      expect(response.body).toEqual({ error: 'Failed to generate user report' });
    });

    it('should handle invalid user id format', async () => {
      // Act
      const response = await request(app).post('/api/user/invalid/report').send({});

      // Assert
      expect(response.status).toBe(400);
      expect(response.body).toEqual({ error: 'Invalid user ID' });
    });
  });
});
