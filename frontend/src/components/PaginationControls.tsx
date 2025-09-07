import React from 'react';

/**
 * PaginationControls Component Props
 */

interface PaginationControlsProps {
  pageNumber: number;
  pageLength: number;
  onPageNumberChange: (value: number) => void;
  onPageLengthChange: (value: number) => void;
  totalRecords: number;
  totalPages: number;
}

/**
 * It's purpose is to provide pagination controls for navigating through paginated data.
 * @param param0 PaginationControlsProps
 * @returns 
 */

export const PaginationControls: React.FC<PaginationControlsProps> = ({
  pageNumber,
  pageLength,
  onPageNumberChange,
  onPageLengthChange,
  totalRecords,
  totalPages,
}) => {
  return (
    <div className="flex items-center gap-4 p-4 bg-white rounded-lg shadow-sm border">
      <div className="flex items-center gap-2">
        <label htmlFor="pageNumber" className="text-sm font-medium text-gray-700">
          Page:
        </label>
        <input
          id="pageNumber"
          type="number"
          min="1"
          max={totalPages}
          value={pageNumber}
          onChange={(e) => onPageNumberChange(Math.max(1, parseInt(e.target.value) || 1))}
          className="w-20 px-2 py-1 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
        />
        <span className="text-sm text-gray-500">of {totalPages}</span>
      </div>

      <div className="flex items-center gap-2">
        <label htmlFor="pageLength" className="text-sm font-medium text-gray-700">
          Per page:
        </label>
        <select
          id="pageLength"
          value={pageLength}
          onChange={(e) => onPageLengthChange(parseInt(e.target.value))}
          className="px-2 py-1 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
        >
          <option value={10}>10</option>
          <option value={20}>20</option>
          <option value={50}>50</option>
          <option value={100}>100</option>
        </select>
      </div>

      <div className="text-sm text-gray-600">
        Total: {totalRecords.toLocaleString()} records
      </div>
    </div>
  );
};