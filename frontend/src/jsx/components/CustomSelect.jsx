import { useState, useRef, useEffect } from "react";
import "./CustomSelect.css";

export default function CustomSelect({
  options,
  placeholder = "Select",
  onChange
}) {
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState(null);
  const ref = useRef(null);

  // закрытие при клике вне&
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (ref.current && !ref.current.contains(e.target)) {
        setOpen(false);
      }
    };

    document.addEventListener("click", handleClickOutside);
    return () => document.removeEventListener("click", handleClickOutside);
  }, []);

  const selectOption = (option) => {
    setSelected(option);
    setOpen(false);
    onChange?.(option);
  };

  return (
    <div className="custom-select" ref={ref}>
      <div
        className={`select-selected ${open ? "active" : ""}`}
        onClick={() => setOpen(!open)}
      >
        {selected || placeholder}
        <span className={`arrow ${open ? "rotate" : ""}`}>▾</span>
      </div>

      {open && (
        <div className="select-options">
          {options.map((opt, i) => (
            <div
              key={i}
              className="select-option"
              onClick={() => selectOption(opt)}
            >
              {opt}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}