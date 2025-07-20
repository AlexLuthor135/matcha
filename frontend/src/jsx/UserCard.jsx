import React from 'react';

const UserCard = ({ user, greyedOut, greyedOutReason, onMouseEnter, onMouseLeave }) => {
  const avatarUrl = `/media/avatars/${user.avatar}`;

  return (
    <div className="relative group w-32 h-32 transition-transform duration-300" onMouseEnter={onMouseEnter} onMouseLeave={onMouseLeave} >
      <div className={`relative w-full h-full rounded-full flex items-center justify-center transition duration-300  ${greyedOut ? 'opacity-40 grayscale' : 'bg-white shadow-md hover:shadow-xl'}`}>
        <div className="w-full h-full rounded-full overflow-hidden">
          <img src={avatarUrl} alt={`${user.student_login}'s avatar`} className="w-full h-full object-cover" />
        </div>
        {greyedOut && greyedOutReason && (
          <div className="absolute bottom-1 left-1/2 -translate-x-1/2 bg-black bg-opacity-60 text-white text-[10px] px-2 py-0.5 rounded pointer-events-none z-10 whitespace-nowrap truncate max-w-[90%] text-center"> {greyedOutReason} </div>
        )}
      </div>
      <div className="absolute left-full ml-2 top-1/2 -translate-y-1/2 max-w-xs bg-black bg-opacity-90 text-white text-xs rounded-lg shadow-lg opacity-0 group-hover:opacity-100 transition-opacity duration-300 p-3 pointer-events-auto whitespace-nowrap" >
        <div className="absolute -left-1 top-1/2 -translate-y-1/2 w-2 h-2 rotate-45 bg-black bg-opacity-90"></div>
        <a href={`https://profile.intra.42.fr/users/${user.student_login}`} target="_blank" rel="noopener noreferrer" className="font-semibold mb-0 hover:underline text-blue-300 block leading-tight"> {user.student_login} </a>
        <p className="leading-none m-0">Rank: {user.quest_rank ?? 'Unranked'}</p>
        <p className="leading-none m-0">Happeer Score: {user.happeer_score}</p>
        <p className="leading-none m-0">Badpeer Score: {user.badpeer_score}</p>
        <p className="leading-none m-0">Weekly Evals: {user.weekly_evals}</p>
      </div>
    </div>
  );
};

export default UserCard;
