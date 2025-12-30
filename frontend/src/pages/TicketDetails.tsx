import { useState, useEffect, useRef, type ChangeEvent } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button } from '../components/Button';
import { Select } from '../components/Select';
import { Textarea } from '../components/Textarea';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { fetchTicketById, updateTicketState, updateTicket } from '../store/slices/ticketsSlice';
import { fetchCommentsByTicketId, createComment } from '../store/slices/commentsSlice';
import { fetchUsers } from '../store/slices/usersSlice';





const normalizeKey = (value?: number | string): string => {
  if (value === undefined || value === null) return '';
  return typeof value === 'string' ? value.toLowerCase() : String(value);
};

type StepStatus = 'complete' | 'current' | 'upcoming';

const isValidUUID = (id: string | null): boolean => {
  if (!id) return false;
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
  return uuidRegex.test(id);
};

const normalizeAssignedList = (userIds: string[] | null | undefined) => {
  if (!userIds) return [];
  return [...userIds].filter((userId) => isValidUUID(userId)).sort();
};

const getUserLabelFromList = (userId: string, users: Array<{id: string; username: string; first_name: string; last_name: string; email: string}>) => {
  const user = users.find(u => u.id === userId);
  if (!user) return userId;
  return `${user.first_name} ${user.last_name} (${user.email})`;
};

const getUserLabel = (userId: string, users: Array<{id: string; username: string; first_name: string; last_name: string; email: string}>) => {
  const user = users.find(u => u.id === userId);
  if (!user) return userId;
  return `${user.first_name} ${user.last_name} (${user.email})`;
};

const normalizeAssignedList = (userIds: string[] | null | undefined) => {
  if (!userIds) return [];
  return [...userIds].filter((userId) => isValidUUID(userId)).sort();
};

const TicketDetails = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { currentTicket, loading: ticketLoading, error: ticketError } = useAppSelector((state) => state.tickets);
  const { comments, loading: commentsLoading, error: commentsError } = useAppSelector((state) => state.comments);
  const { users, loading: usersLoading } = useAppSelector((state) => state.users);

  const [newComment, setNewComment] = useState('');
  const [assignedTo, setAssignedTo] = useState<string[]>([]);
  const [priority, setPriority] = useState('low');
  const [state, setState] = useState('open');
  const [description, setDescription] = useState('');
  const [hasChanges, setHasChanges] = useState(false);
  const [assigneeOpen, setAssigneeOpen] = useState(false);
  const assigneeRef = useRef<HTMLDivElement | null>(null);

  // Fetch ticket data based on ID
  useEffect(() => {
    if (id) {
      dispatch(fetchTicketById(id));
      dispatch(fetchCommentsByTicketId(id));
      dispatch(fetchUsers());
    }
  }, [id, dispatch]);

  useEffect(() => {
    const handleOutsideClick = (event: MouseEvent) => {
      if (!assigneeRef.current) return;
      if (event.target instanceof Node && !assigneeRef.current.contains(event.target)) {
        setAssigneeOpen(false);
      }
    };
    document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, []);

  // Update form fields when currentTicket changes
  useEffect(() => {
    if (currentTicket) {
      setAssignedTo(normalizeAssignedList(currentTicket.assigned_to));
      const priorityValue = normalizeKey(currentTicket.priority);
      const validPriority = ['low', 'medium', 'high', 'critical'].includes(priorityValue) ? priorityValue : 'low';
      setPriority(validPriority);
      const stateValue = normalizeKey(currentTicket.state);
      let validState = 'open';
      if (['open', 'pending', 'resolved', 'closed', 'cancelled'].includes(stateValue)) {
        validState = stateValue;
      } else if (stateValue === 'cancel') {
        validState = 'cancelled';
      }
      setState(validState);
      setDescription(currentTicket.description || '');
      setHasChanges(false);
    }
  }, [currentTicket]);

  // Check for changes
  useEffect(() => {
    if (currentTicket) {
      const changed =
        normalizeAssignedList(assignedTo).join('|') !== normalizeAssignedList(currentTicket.assigned_to).join('|') ||
        priority !== normalizeKey(currentTicket.priority) ||
        state !== normalizeKey(currentTicket.state) ||
        description !== (currentTicket.description || '');
      setHasChanges(changed);
    }
  }, [assignedTo, priority, state, description, currentTicket]);

  const handleAddComment = () => {
    if (newComment.trim() === '' || !id) return;

    dispatch(createComment({
      ticket_id: id,
      description: newComment
    })).then(() => {
      setNewComment("");
    });
  };

  const handleCancelTicket = async () => {
    if (!id || ticketLoading) return;
    await dispatch(updateTicketState({ id, state: 'cancelled' }));
    dispatch(fetchTicketById(id));
  };
  const handleResolveTicket = async () => {
    if (!id || ticketLoading) return;
    await dispatch(updateTicketState({ id, state: 'resolved' }));
    dispatch(fetchTicketById(id));
  };
  const handleCancelClick = () => {
    void handleCancelTicket();
  };
  const handleResolveClick = () => {
    void handleResolveTicket();
  };

  const handleRemoveAssignee = (userId: string) => {
    setAssignedTo((prev) => prev.filter((id) => id !== userId));
  };

  const handleToggleAssignee = (userId: string) => {
    setAssignedTo((prev) => (
      prev.includes(userId) ? prev.filter((id) => id !== userId) : [...prev, userId]
    ));
  };

  const handlePriorityChange = (e: ChangeEvent<HTMLSelectElement>) => {
    setPriority(e.target.value);
  };

  const handleStateChange = (e: ChangeEvent<HTMLSelectElement>) => {
    setState(e.target.value);
  };

  const handleDescriptionChange = (e: ChangeEvent<HTMLTextAreaElement>) => {
    setDescription(e.target.value);
  };

  const handleUpdateTicket = async () => {
    if (!id || !hasChanges || ticketLoading) return;

    await dispatch(updateTicket({ 
      id, 
      assignedTo: assignedTo,
      priority: priority,
      state: state,
      description: description
    }));
    
    dispatch(fetchTicketById(id));
  };

  if (ticketLoading || commentsLoading) {
    return (
      <div className="container mx-auto px-4 py-8 flex justify-center items-center h-64">
        <p className="text-lg">Loading ticket details...</p>
      </div>
    );
  }

  if (ticketError) {
    return (
      <div className="container mx-auto px-4 py-8 text-red-600">
        Error: {ticketError}
      </div>
    );
  }

  if (commentsError) {
    return (
      <div className="container mx-auto px-4 py-8 text-red-600">
        Error: {commentsError}
      </div>
    );
  }

  if (!currentTicket) {
    return (
      <div className="container mx-auto px-4 py-8">
        <p className="text-lg">Ticket not found</p>
      </div>
    );
  }

  const lifecycleSteps = [
    { key: 1, label: 'Open', aliases: [1, 'open'] },
    { key: 2, label: 'Pending', aliases: [2, 'pending'] },
    { key: 3, label: 'Resolved', aliases: [3, 'resolved'] },
    { key: 4, label: 'Closed', aliases: [4, 'closed'] },
    { key: 5, label: 'Cancelled', aliases: [5, 'cancelled', 'cancel'] },
  ];
  const activeKey = currentTicket.state;
  const activeIndex = lifecycleSteps.findIndex((item) => item.aliases.includes(activeKey));
  const isCancelable = activeKey !== 5 && activeKey !== 4;
  const isResolvable = activeKey !== 3 && activeKey !== 4 && activeKey !== 5;

  const userOptions = users.map(user => ({
    value: user.id,
    label: `${user.first_name} ${user.last_name} (${user.email})`
  }));

  const priorityOptions = [
    { value: 'critical', label: 'Critical' },
    { value: 'high', label: 'High' },
    { value: 'medium', label: 'Medium' },
    { value: 'low', label: 'Low' }
  ];

  const stateOptions = [
    { value: 'open', label: 'Open' },
    { value: 'pending', label: 'Pending' },
    { value: 'resolved', label: 'Resolved' },
    { value: 'closed', label: 'Closed' },
    { value: 'cancelled', label: 'Cancelled' }
  ];

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-4">
        <button
          onClick={() => navigate('/tickets')}
          className="text-blue-600 hover:text-blue-800 flex items-center"
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-1" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M9.707 14.707a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 1.414L7.414 9H15a1 1 0 110 2H7.414l2.293 2.293a1 1 0 010 1.414z" clipRule="evenodd" />
          </svg>
          Back to Tickets
        </button>
      </div>

      <div className="bg-white shadow-md rounded-lg p-6 dark:bg-gray-800 transition-colors duration-200">
        <div className="flex items-start justify-between mb-4">
          <h1 className="text-2xl font-bold dark:text-white">{currentTicket.title}</h1>
          <div className="flex items-center gap-2">
            <Button
              label="Resolve"
              onClick={handleResolveClick}
              disabled={!isResolvable || ticketLoading}
            />
            <Button
              label="Update"
              onClick={handleUpdateTicket}
              disabled={!hasChanges || ticketLoading}
            />
            <Button
              label="Cancel Ticket"
              onClick={handleCancelClick}
              disabled={!isCancelable || ticketLoading}
            />
          </div>
        </div>

        <div className="mb-8">
          <div className="flex items-center gap-2 md:gap-4">
            {lifecycleSteps.map((step, index) => {
              const status: StepStatus = activeIndex === -1
                ? 'upcoming'
                : index < activeIndex
                  ? 'complete'
                  : index === activeIndex
                    ? 'current'
                    : 'upcoming';
              const isLast = index === lifecycleSteps.length - 1;
              const dotColor = status === 'complete'
                ? 'bg-blue-600'
                : status === 'current'
                  ? 'bg-blue-100 ring-2 ring-blue-600 dark:bg-blue-900/60'
                  : 'bg-gray-200 dark:bg-gray-700';
              const connectorColor = status === 'complete'
                ? 'bg-blue-600'
                : status === 'current'
                  ? 'bg-blue-200 dark:bg-blue-900/60'
                  : 'bg-gray-200 dark:bg-gray-700';

              return (
                <div key={step.key} className="flex items-center flex-1 min-w-[80px]">
                  <div className="flex flex-col items-center w-full">
                    <div className="flex items-center justify-center w-full">
                      <div className={`w-4 h-4 rounded-full transition-colors duration-200 ${dotColor}`} aria-hidden />
                    </div>
                    <span className="mt-2 text-xs font-medium text-gray-700 dark:text-gray-200 text-center">{step.label}</span>
                  </div>
                  {!isLast && <div className={`h-0.5 flex-1 mx-2 ${connectorColor}`} />}
                </div>
              );
            })}
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 mb-6">
          <div>
            <p className="text-sm text-gray-500 dark:text-gray-400">Created By</p>
            <p className="font-medium dark:text-white">
              {currentTicket.creator ? `${currentTicket.creator.first_name} ${currentTicket.creator.last_name} (${currentTicket.creator.email})` : 'Unknown'}
            </p>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Assigned To</label>
            <div className="relative mt-1" ref={assigneeRef}>
              <button
                type="button"
                onClick={() => setAssigneeOpen((open) => !open)}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-left shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700"
                disabled={ticketLoading || usersLoading}
                aria-expanded={assigneeOpen}
              >
                <div className="flex flex-wrap gap-2 items-center">
                  {assignedTo.length === 0 && (
                    <span className="text-sm text-gray-400 dark:text-gray-400">Select users</span>
                  )}
                  {assignedTo.map((userId) => (
                    <span
                      key={userId}
                      className="inline-flex items-center gap-1 rounded-full bg-slate-200 text-slate-800 text-xs px-3 py-1 dark:bg-slate-600 dark:text-white"
                    >
                      {getUserLabelFromList(userId, users)}
                      <button
                        type="button"
                        onClick={(event) => {
                          event.stopPropagation();
                          handleRemoveAssignee(userId);
                        }}
                        className="ml-1 rounded-full px-1 text-slate-600 hover:text-slate-900 dark:text-slate-200 dark:hover:text-white"
                        aria-label={`Remove ${getUserLabelFromList(userId, users)}`}
                      >
                        x
                      </button>
                    </span>
                  ))}
                </div>
                <span className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 dark:text-gray-300">
                  {assigneeOpen ? '^' : 'v'}
                </span>
              </button>
              {assigneeOpen && (
                <div className="absolute z-20 mt-2 w-full rounded-md border border-gray-300 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800">
                  <ul className="max-h-64 overflow-auto py-1">
                    {userOptions.map((option) => {
                      const isSelected = assignedTo.includes(option.value);
                      return (
                        <li key={option.value}>
                          <button
                            type="button"
                            onClick={() => handleToggleAssignee(option.value)}
                            className={`w-full px-4 py-2 text-left text-sm ${isSelected ? 'bg-slate-600 text-white' : 'text-gray-900 hover:bg-gray-100 dark:text-gray-100 dark:hover:bg-gray-700'}`}
                          >
                            {option.label}
                          </button>
                        </li>
                      );
                    })}
                  </ul>
                </div>
              )}
            </div>
          </div>
          <div>
            <Select
              label="Priority"
              name="priority"
              value={priority}
              options={priorityOptions}
              onChange={handlePriorityChange}
              disabled={ticketLoading}
              showUnassigned={false}
            />
          </div>
          <div>
            <Select
              label="State"
              name="state"
              value={state}
              options={stateOptions}
              onChange={handleStateChange}
              disabled={true}
              showUnassigned={false}
            />
          </div>
        </div>

        <div className="mb-6">
          <Textarea
            label="Description"
            name="description"
            value={description}
            onChange={handleDescriptionChange}
            disabled={ticketLoading}
            rows={4}
          />
        </div>

        <div className="mb-6">
          <p className="text-sm text-gray-500 mb-2 dark:text-gray-400">Skills</p>
          <div className="flex flex-wrap gap-2">
            {/* {ticket.skills.map((skill, index) => (
              <span key={index} className="bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded">{skill}</span>
            ))} */}
          </div>
        </div>

        <div className="border-t pt-6 dark:border-gray-700">
          <h2 className="text-lg font-semibold mb-4 dark:text-white">Comments</h2>

          <div className="space-y-4 mb-6">
            {comments.map((comment) => (
              <div key={comment.id} className="bg-gray-50 p-4 rounded dark:bg-gray-700">
                <div className="flex justify-between mb-2">
                  <p className="font-medium dark:text-white">
                    {comment.creator ? `${comment.creator.first_name} ${comment.creator.last_name} (${comment.creator.email})` : 'Unknown'}
                  </p>
                  <p className="text-sm text-gray-500 dark:text-gray-400">{new Date(comment.created_at).toLocaleString()}</p>
                </div>
                <p className="dark:text-gray-300">{comment.description}</p>
              </div>
            ))}
          </div>

          <div className="mt-4">
            <div className="flex items-center justify-between mb-2">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300" htmlFor="comment-input">
                Add a comment
              </label>
              <Button label="Post" onClick={handleAddComment} />
            </div>
            <input
              id="comment-input"
              type="text"
              name="comment"
              value={newComment}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setNewComment(e.target.value)}
              required
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:placeholder-gray-400"
            />
          </div>
        </div>

      </div>
    </div>
  );
};
export default TicketDetails;
