import React, { useEffect, useState } from 'react';
import axiosInstance from './AxiosInstance';
import '../css/ProfilePage.css';
import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import Loading from './LoadingScreen';
import { getBlockedDuration } from './BlockedDuration';

export default function ProfilePage() {
  const [teams, setTeams] = useState(null);
  const [userInfo, setUserInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [blockedDurationText, setBlockedDurationText] = useState('');
  const [blockPassCooldownText, setBlockPassCooldownText] = useState('');

  useEffect(() => {
    loadTeams();
  }, []);

  useEffect(() => {
    if (!userInfo?.blocked_until) {
      setBlockedDurationText('Not currently blocked');
      return;
    }
    const updateDuration = () => {
      const text = getBlockedDuration(userInfo.blocked_until);
      setBlockedDurationText(text);
    };
    updateDuration();
    const intervalId = setInterval(updateDuration, 1000);
    return () => clearInterval(intervalId);
  }, [userInfo?.blocked_until]);

  useEffect(() => {
    if (!userInfo?.block_pass_used_at) {
      setBlockPassCooldownText('Available ✅');
      return;
    }
    const updateCooldown = () => {
      const text = getBlockedDuration(userInfo.block_pass_used_at, 7 * 24 * 60 * 60 * 1000);
      setBlockPassCooldownText(text);
    };
  
    updateCooldown();
    const intervalId = setInterval(updateCooldown, 1000);
    return () => clearInterval(intervalId);
  }, [userInfo?.block_pass_used_at]);
  
  const loadTeams = async () => {
    setLoading(true);
    try {
      const response = await axiosInstance.get('backend/intra/teams/');
      setTeams(response.data.teams);
      setUserInfo(response.data.user_info);
    } catch (error) {
      const errorMsg = error.response?.data?.error || "Failed to load teams.";
      toast.error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateEvaluation = async (teamId, projectName) => {
    setLoading(true);
    try {
      await axiosInstance.post('backend/intra/teams/assign/', { team_id: teamId, project_name: projectName });
      toast.success("Evaluation created successfully!");
      loadTeams();
    } catch (error) {
      const errorMsg = error.response?.data?.error || "An unexpected error occurred.";
      toast.error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const handleResetBlocked = async () => {
    try {
      const response = await axiosInstance.post('backend/api/user/reset_blocked/');
      toast.success(response.data.message);
      loadTeams();
    } catch (error) {
      const errorMsg = error.response?.data?.error || "Failed to reset blocked status.";
      toast.error(errorMsg);
    }
  };

  const handleSetBlocked = async () => {
    try {
      const response = await axiosInstance.post('backend/api/user/set_blocked/');
      toast.success(response.data.message);
      loadTeams();
    } catch (error) {
      const errorMsg = error.response?.data?.error || "Failed to block user.";
      toast.error(errorMsg);
    }
  };
  
  
  const renderEvaluationCount = (team) => {
    if (team.correction_number === null) {
      return (
        <p>
          Evaluations: <span style={{ color: '#999' }}>(no total available)</span>
        </p>
      );
    }
    return (
      <p>
        Evaluations: {team.scale_teams}/{team.correction_number}
      </p>
    );
  };
  
  const renderEvaluationAction = (team) => {
    if (team.correction_number === null) {
      return (
        <>
          <p>No evaluations occurred yet.</p>
          <button onClick={() => handleCreateEvaluation(team.id, team.project_name)}>
            Create Evaluation
          </button>
        </>
      );
    }
    if (team.scale_teams < team.correction_number) {
      return (
        <button onClick={() => handleCreateEvaluation(team.id, team.project_name)}>
          Create Evaluation
        </button>
      );
    }
    return <p>Evaluations Full</p>;
  };
  
  if (loading) {
    return <Loading message="Gathering info..." />;
  }  

  return (
    <div className="profile-page">
      {userInfo && (
        <div className="user-info-box">
          <h2>Your Stats</h2>
          <p>Weekly Evaluations: {userInfo.weekly_evals ?? 0}/3</p>
          <p className="info-line">
            <span>Block Pass Cooldown: </span>
            <span>{blockPassCooldownText}</span>
          </p>
          <p>Available Overdrive: {userInfo.available_pass ? 'Active ✅' : 'Inactive ❌'}</p>
          {userInfo.blocked_until ? (
            <p>Unavailable for: {blockedDurationText}</p>
          ) : (
            <p>Not currently blocked</p>
          )}
          <div className="user-info-buttons">
            <button className="user-action-button" onClick={() => {
              const confirmed = window.confirm(
                "Are you sure you want to reset your blocked status? Without the Block Pass, you won't be able to block yourself again."
              );
              if (!confirmed) {
                return;
              }
              handleResetBlocked();
            }} disabled={loading}>Set as available</button>
            <button className="user-action-button" onClick={() => {
                const confirmed = window.confirm(
                  "Are you sure you want to set yourself as unavailable? It can be done once per week."
                );
                if (!confirmed) {
                  return;
                }
                handleSetBlocked();
              }}
              disabled={loading}
            >
              Set as unavailable
            </button>
          </div>
        </div>
      )}
      <div className="teams-section">
        <h1 style={{ color: "#ffffff" }}>You're logged in!</h1>
        <div className="teams-container">
          {teams && teams.length > 0 ? (
            teams.map((team) => (
              <div key={team.id} className="team-box">
                <h2>{team.name}</h2>
                <p>Project: {team.project_name}</p>
                {renderEvaluationCount(team)}
                {renderEvaluationAction(team)}
              </div>
            ))
          ) : (
            <p style={{ color: "#ffffff" }}>No teams available.</p>
          )}
        </div>
      </div>
    </div>
  );
}
