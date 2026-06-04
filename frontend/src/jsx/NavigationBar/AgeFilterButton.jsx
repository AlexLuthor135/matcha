import { useEffect, useRef, useState } from "react";
import "./AgeFilterButton.css";

const MIN_AGE = 18;
const MAX_AGE = 80;

function clampAge(value) {
  const numericValue = Number(value);

  if (Number.isNaN(numericValue)) {
    return MIN_AGE;
  }

  return Math.min(MAX_AGE, Math.max(MIN_AGE, numericValue));
}

export default function AgeFilterButton() {
  const [isOpen, setIsOpen] = useState(false);
  const [ageRange, setAgeRange] = useState({
    min: MIN_AGE,
    max: MAX_AGE,
  });
  const filterRef = useRef(null);

  useEffect(() => {
    const handlePointerDown = (event) => {
      if (!filterRef.current?.contains(event.target)) {
        setIsOpen(false);
      }
    };

    document.addEventListener("pointerdown", handlePointerDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
    };
  }, []);

  const updateMinAge = (value) => {
    setAgeRange((currentRange) => ({
      ...currentRange,
      min: Math.min(clampAge(value), currentRange.max),
    }));
  };

  const updateMaxAge = (value) => {
    setAgeRange((currentRange) => ({
      ...currentRange,
      max: Math.max(clampAge(value), currentRange.min),
    }));
  };

  return (
    <div className="age-filter" ref={filterRef}>
      <button
        className="age-filter-button"
        type="button"
        aria-expanded={isOpen}
        aria-haspopup="true"
        onClick={() => setIsOpen((currentIsOpen) => !currentIsOpen)}
      >
        Filter
      </button>

      {isOpen ? (
        <div className="age-filter-menu" role="menu">
          <div className="age-filter-header">
            <h3>Age range</h3>
            <span>
              {ageRange.min}-{ageRange.max}
            </span>
          </div>

          <div className="age-filter-fields">
            <label>
              <span>From</span>
              <input
                type="number"
                min={MIN_AGE}
                max={ageRange.max}
                value={ageRange.min}
                onChange={(event) => updateMinAge(event.target.value)}
              />
            </label>

            <label>
              <span>To</span>
              <input
                type="number"
                min={ageRange.min}
                max={MAX_AGE}
                value={ageRange.max}
                onChange={(event) => updateMaxAge(event.target.value)}
              />
            </label>
          </div>

          <div className="age-filter-sliders">
            <input
              type="range"
              min={MIN_AGE}
              max={MAX_AGE}
              value={ageRange.min}
              aria-label="Minimum age"
              onChange={(event) => updateMinAge(event.target.value)}
            />
            <input
              type="range"
              min={MIN_AGE}
              max={MAX_AGE}
              value={ageRange.max}
              aria-label="Maximum age"
              onChange={(event) => updateMaxAge(event.target.value)}
            />
          </div>
        </div>
      ) : null}
    </div>
  );
}
