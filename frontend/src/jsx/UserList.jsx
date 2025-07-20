import React, { useEffect, useRef, useState } from 'react';
import axiosInstance from './AxiosInstance';
import UserCard from './UserCard';
import Loading from './LoadingScreen';
import BackButton from './BackButton';
import { getBlockedDuration } from './BlockedDuration';


const FILTERS = {
  ALL: 'all',
  TIME: 'time',
  WEEKLY: 'weekly limit',
  EVAL: 'evaluating',
};

const generateGridPositions = (count) => {
  const cols = Math.ceil(Math.sqrt(count));
  const rows = Math.ceil(count / cols);
  const positions = [];

  for (let r = 1; r <= rows; r++) {
    for (let c = 1; c <= cols; c++) {
      positions.push({
        top: `${(r * 100) / (rows + 1)}%`,
        left: `${(c * 100) / (cols + 1)}%`,
      });
    }
  }

  return positions.slice(0, count);
};

const UserList = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [showGreyedOut, setShowGreyedOut] = useState(false);
  const [blockFilter, setBlockFilter] = useState(FILTERS.ALL);
  const [positions, setPositions] = useState([]);
  const [focusedCardId, setFocusedCardId] = useState(null);
  const hoverTimeoutRef = useRef(null);
  const [, forceRerender] = useState(0);

  useEffect(() => {
    fetchAllUsers();
  }, []);

  useEffect(() => {
    const interval = setInterval(() => {
      forceRerender((n) => n + 1);
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  const fetchAllUsers = async (url = '/backend/admins/users/', collected = []) => {
    try {
      setLoading(true);
      const response = await axiosInstance.get(url);
      if (!Array.isArray(response.data.results)) return;

      const combined = [...collected, ...response.data.results];

      if (response.data.next) {
        const nextPage = new URL(response.data.next);
        const pageParam = nextPage.searchParams.get('page');
        const correctedUrl = `/backend/admins/users/?page=${pageParam}`;
        await fetchAllUsers(correctedUrl, combined);
      } else {
        setUsers(combined);
        setPositions(generateGridPositions(combined.length));
      }
    } catch (err) {
      console.error('Failed to load users:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleMouseEnter = (id) => {
    if (hoverTimeoutRef.current) clearTimeout(hoverTimeoutRef.current);
    setFocusedCardId(id);
  };

  const handleMouseLeave = () => {
    hoverTimeoutRef.current = setTimeout(() => {
      setFocusedCardId(null);
    }, 300);
  };

  const totalUsersCount = users.length;

  const filteredUsersCount = users.reduce((count, user) => {
    const hasEval = user.has_eval;
    const isTimeBlocked = user.blocked_until && new Date(user.blocked_until) > new Date();
    const isEvalBlocked = user.weekly_evals >= 3;

    if (!showGreyedOut) return count;

    const showForAll = blockFilter === FILTERS.ALL && (isTimeBlocked || isEvalBlocked || hasEval);
    const showForTime = blockFilter === FILTERS.TIME && isTimeBlocked;
    const showForEval = blockFilter === FILTERS.WEEKLY && isEvalBlocked;
    const showForEvaluating = blockFilter === FILTERS.EVAL && hasEval;

    return count + (showForAll || showForTime || showForEval || showForEvaluating ? 1 : 0);
  }, 0);

  return (
    <div className="relative w-full h-screen overflow-hidden bg-transparent">
      {loading && <Loading text="Gathering user data..." />}

      <div className="absolute top-4 right-4 z-50 flex flex-wrap items-center gap-2 px-3 py-2">
      <BackButton />
        <button
          onClick={() => setShowGreyedOut(!showGreyedOut)}
          className="text-sm text-black hover:underline"
        >
          {showGreyedOut ? 'Hide' : 'Show'} Blocked Students
        </button>

        {showGreyedOut && (
          <>
            {Object.values(FILTERS).map((type) => (
              <button
                key={type}
                onClick={() => setBlockFilter(type)}
                className={`text-sm px-1 py-0.5 ${
                  blockFilter === type
                    ? 'font-semibold underline text-black'
                    : 'text-gray-700 hover:underline'
                }`}
              >
                {type.charAt(0).toUpperCase() + type.slice(1)}
              </button>
            ))}
            <span className="text-sm font-medium text-black px-2 py-1 bg-white/80 rounded shadow-sm ml-1">
              Showing {filteredUsersCount} of {totalUsersCount}
            </span>
          </>
        )}
      </div>
      <div className="absolute inset-0 overflow-auto">
        <div className="relative min-h-[120vh]">
          {users.map((user, index) => {
            const pos = positions[index] || { top: '50%', left: '50%' };
            const hasEval = user.has_eval;
            const isTimeBlocked = user.blocked_until && new Date(user.blocked_until) > new Date();
            const isEvalBlocked = user.weekly_evals >= 3;
            const isFocused = focusedCardId === user.id;
          
            let greyedOut = false;
            let greyedOutReason = null;
          
            if (showGreyedOut) {
              const showForAll =
                blockFilter === FILTERS.ALL && (isTimeBlocked || isEvalBlocked || hasEval);
              const showForTime = blockFilter === FILTERS.TIME && isTimeBlocked;
              const showForEval = blockFilter === FILTERS.WEEKLY && isEvalBlocked;
              const showForEvaluating = blockFilter === FILTERS.EVAL && hasEval;
            
              greyedOut = showForAll || showForTime || showForEval || showForEvaluating;
            
              if (isTimeBlocked) {
                greyedOutReason = `(${getBlockedDuration(user.blocked_until)})`;
              } else if (isEvalBlocked) {
                greyedOutReason = `Weekly Limit (${user.weekly_evals})`;
              } else if (hasEval) {
                greyedOutReason = 'Evaluating';
              }
            }
          
            return (
              <div
                key={user.id}
                className={`absolute ${isFocused ? 'z-50' : 'z-10'}`}
                style={{
                  top: pos.top,
                  left: pos.left,
                  transform: 'translate(-50%, -50%)',
                }}
              >
                <UserCard
                  user={user}
                  greyedOut={greyedOut}
                  greyedOutReason={greyedOutReason}
                  onMouseEnter={() => handleMouseEnter(user.id)}
                  onMouseLeave={handleMouseLeave}
                />
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default UserList;
