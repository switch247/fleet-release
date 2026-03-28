sample questions
questions.md: Record all your questions during the process of understanding the Prompt. questions.md: Record all questions you have when understanding the business logic in the Prompt. Key points: Unclear business-level aspects such as business processes, business rules, data relationships, and boundary conditions. Format: Question + Your understanding/hypothesis + Solution
Example: Required Document Description: Business Logic Questions Log:
Inventory rollback logic after order cancellation
Question: Prompt mentioned "users can cancel orders," but did not specify how inventory is handled after a paid order is canceled.
My Understanding: Canceled paid orders should roll back inventory immediately; unpaid orders cancel automatically upon timeout.
Solution: Implemented an order state machine where cancellation triggers an inventory rollback event.
Inheritance of user permissions
Question: Prompt includes "Admin" and "Super Admin," but did not clarify the differences in permission scope.
My Understanding: Super Admin can manage all data; regular Admins can only manage data they created.
Solution: Added a scope field to the permissions table to distinguish permission ranges.
Data deletion: physical or logical delete
Question: Prompt did not specify if "deleting a user" is a hard delete or a marked delete.
My Understanding: Considering data auditing requirements, a logical delete (soft delete) will be used.
Solution: Added a deleted_at field to all tables.