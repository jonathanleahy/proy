import { Router, Request, Response } from 'express';
import { UserService } from '../services/userService';
import { PeopleService } from '../services/peopleService';
import { ReportRequest } from '../types';

/**
 * Creates user router with dependency injection
 * @param userService - Optional UserService instance (defaults to new instance)
 * @param peopleService - Optional PeopleService instance (defaults to new instance)
 */
export function createUserRouter(
  userService: UserService = new UserService(),
  peopleService: PeopleService = new PeopleService()
): Router {
  const router = Router();

  /**
   * Endpoint 1: Simple - Get user by ID
   * GET /api/user/:id
   */
  router.get('/user/:id', async (req: Request, res: Response) => {
    try {
      const userId = parseInt(req.params.id, 10);

      // Validate user ID
      if (isNaN(userId)) {
        return res.status(400).json({ error: 'Invalid user ID' });
      }

      const user = await userService.getUser(userId);
      return res.status(200).json(user);
    } catch (error: any) {
      if (error.message === 'User not found') {
        return res.status(404).json({ error: error.message });
      }
      return res.status(500).json({ error: error.message });
    }
  });

  /**
   * Endpoint 2: Medium complexity - Get user summary with posts
   * GET /api/user/:id/summary
   */
  router.get('/user/:id/summary', async (req: Request, res: Response) => {
    try {
      const userId = parseInt(req.params.id, 10);

      // Validate user ID
      if (isNaN(userId)) {
        return res.status(400).json({ error: 'Invalid user ID' });
      }

      const summary = await userService.getUserSummary(userId);
      return res.status(200).json(summary);
    } catch (error: any) {
      return res.status(500).json({ error: error.message });
    }
  });

  /**
   * Endpoint 3: Complex - Get comprehensive user report
   * POST /api/user/:id/report
   */
  router.post('/user/:id/report', async (req: Request, res: Response) => {
    try {
      const userId = parseInt(req.params.id, 10);

      // Validate user ID
      if (isNaN(userId)) {
        return res.status(400).json({ error: 'Invalid user ID' });
      }

      const options: ReportRequest = req.body || {};
      const report = await userService.getUserReport(userId, options);
      return res.status(200).json(report);
    } catch (error: any) {
      return res.status(500).json({ error: error.message });
    }
  });

  /**
   * Endpoint 4: Person lookup - Find person by surname and DOB
   * GET /api/person?surname=Thompson&dob=1985-03-15
   * Returns single person object
   */
  router.get('/person', async (req: Request, res: Response) => {
    try {
      const { surname, dob } = req.query;

      // Validate required parameters
      if (!surname || !dob) {
        return res.status(400).json({
          error: 'Missing required parameters',
          message: 'Both surname and dob are required',
        });
      }

      // Validate dob format (basic check)
      if (typeof dob !== 'string' || !/^\d{4}-\d{2}-\d{2}$/.test(dob)) {
        return res.status(400).json({
          error: 'Invalid date format',
          message: 'dob must be in format YYYY-MM-DD',
        });
      }

      const person = await peopleService.findPerson(surname as string, dob as string);

      if (!person) {
        return res.status(404).json({
          error: 'Person not found',
          message: `No person found with surname "${surname}" and dob "${dob}"`,
        });
      }

      return res.status(200).json(person);
    } catch (error: any) {
      return res.status(500).json({ error: error.message });
    }
  });

  /**
   * Endpoint 5: People search - Search by surname OR dob
   * GET /api/people?surname=Thompson
   * GET /api/people?dob=1985-03-15
   * Returns array of matching people
   */
  router.get('/people', async (req: Request, res: Response) => {
    try {
      const { surname, dob } = req.query;

      // Validate at least one parameter
      if (!surname && !dob) {
        return res.status(400).json({
          error: 'Missing required parameters',
          message: 'At least one of surname or dob is required',
        });
      }

      // Validate dob format if provided
      if (dob && (typeof dob !== 'string' || !/^\d{4}-\d{2}-\d{2}$/.test(dob))) {
        return res.status(400).json({
          error: 'Invalid date format',
          message: 'dob must be in format YYYY-MM-DD',
        });
      }

      const people = await peopleService.findPeople(
        surname as string | undefined,
        dob as string | undefined
      );

      return res.status(200).json(people);
    } catch (error: any) {
      return res.status(500).json({ error: error.message });
    }
  });

  return router;
}
