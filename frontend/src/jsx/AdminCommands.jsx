import React, { useState } from 'react';
import { toast } from 'react-toastify';
import axiosInstance from './AxiosInstance';
import BackButton from './BackButton';

const commands = [
    { label: 'Reset Locations', command: 'populate_locations', description: 'Resets all locations and sets those at school as active.' },
    { label: 'Reset Friends', command: 'populate_friends', description: 'Re-populates past evaluators (friends) for each user.' },
    { label: 'Reset Ranks', command: 'populate_ranks', description: 'Re-populates each user\'s rank.'},
    { label: 'Add New Cohort', command: 'populate_cohort', description: 'Adds new/missing students to the DB with all the values being set.' },
    { label: 'Populate Avatars', command: 'populate_avatars', description: 'Populates missing avatars for existing users in DB.' },
    { label: 'Reset "has_eval"', command: 'reset_has_eval', description: 'Resets "has_eval" field for all users.' },
    { label: 'Delete 1000 evaluations points', command: 'reset_eval_points', description: 'Removes 1000 eval points from each user in DB on Intra.' },
    { label: 'Update pending Evals', command: 'update_pending_evals', description: 'Checks all pending evals for missed webhooks. (should run regularly)' },
]

const AdminCommandPanel = () => {
    const [loadingCommand, setLoadingCommand] = useState(null);
  
    const handleCommand = async (commandName) => {
      setLoadingCommand(commandName);
      try {
        const response = await axiosInstance.post('/backend/admins/commands/', {
          command: commandName,
        });
        toast.success(`Success: ${response.data.message || 'Command executed'}`);
      } catch (error) {
        console.error('Command failed:', error);
        toast.error(`Error: ${error.response?.data?.error || 'Command failed'}`);
      } finally {
        setLoadingCommand(null);
      }
    };
  
    return (
      <div className="max-w-md mx-auto mt-10 px-4">
        <BackButton />
      <div className="p-6 bg-white rounded shadow space-y-4 mt-4">
        <h2 className="text-lg font-semibold text-gray-700 text-center">Admin Commands</h2>
        <div className="flex flex-col space-y-3 items-center">
          {commands.map((cmd) => (
            <div key={cmd.command} className="relative group inline-block">
              <button
                onClick={() => {
                  const confirmed = window.confirm(`Are you sure you want to run "${cmd.label}"?`);
                  if (confirmed) {
                    handleCommand(cmd.command);
                  }
                }}
                disabled={loadingCommand !== null}
                className={`w-64 px-5 py-2 rounded text-sm font-medium transition ${
                  loadingCommand === cmd.command
                    ? 'bg-blue-300 cursor-wait text-blue-500'
                    : 'bg-blue-500 hover:bg-blue-600 text-blue-500'
                }`}
              >
                {loadingCommand === cmd.command ? 'Running...' : cmd.label}
              </button>
              <div className="absolute left-full ml-2 top-1/2 -translate-y-1/2 hidden group-hover:block w-64 px-3 py-2 text-xs text-white bg-gray-800 rounded shadow-lg z-10 text-center">
                {cmd.description}
              </div>
            </div>
          ))}
        </div>
      </div>
      </div>
    );
};
  
  export default AdminCommandPanel;
