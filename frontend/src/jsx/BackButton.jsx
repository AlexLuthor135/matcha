import { useNavigate } from 'react-router-dom';

const BackButton = () => {
  const navigate = useNavigate();

  return (
    <button
      onClick={() => navigate(-1)}
      className="inline-flex items-center px-4 py-2 text-sm font-medium text-blue-500 border border-blue-500 rounded hover:bg-blue-50 transition"
    >
      ← Back
    </button>
  );
};

export default BackButton;
