interface Option {
  value: string;
  label: string;
}

interface SelectProps {
  label: string;
  name: string;
  value: string | string[];
  options: Option[];
  onChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  className?: string;
  disabled?: boolean;
  showUnassigned?: boolean;
  multiple?: boolean;
}

export const Select = ({
  label,
  name,
  value,
  options,
  onChange,
  className = '',
  disabled = false,
  showUnassigned = true,
  multiple = false,
}: SelectProps) => {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">{label}</label>
      <select
        name={name}
        value={value}
        onChange={onChange}
        disabled={disabled}
        multiple={multiple}
        className={`mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white ${className}`}
      >
        {showUnassigned && !multiple && <option value="">Unassigned</option>}
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </div>
  );
};
