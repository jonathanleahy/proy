export const UserService = jest.fn().mockImplementation(() => ({
  getUser: jest.fn(),
  getUserSummary: jest.fn(),
  getUserReport: jest.fn(),
}));
