import '../css/Stars.css';

const Stars = () => {
  const generateStars = () => {
    const starCount = 100;
    const starArray = [];
    
    for (let i = 0; i < starCount; i++) {
      const size = Math.random() * 3 + 1;
      const positionX = Math.random() * 100;
      const positionY = Math.random() * 100;
      const animationDuration = Math.random() * 5 + 5;

      starArray.push(
        <div
          key={i}
          className="star"
          style={{
            width: `${size}px`,
            height: `${size}px`,
            left: `${positionX}%`,
            top: `${positionY}%`,
            animationDuration: `${animationDuration}s`,
          }}
        />
      );
    }

    return starArray;
  };

  return <div className="stars">{generateStars()}</div>;
};

export default Stars;
