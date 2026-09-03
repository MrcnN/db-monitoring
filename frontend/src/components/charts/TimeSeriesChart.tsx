import React from 'react';

interface DataPoint {
  time: string;
  value: number;
}

interface TimeSeriesChartProps {
  title: string;
  data: DataPoint[];
  unit?: string;
  color?: string; // hex or tailwind stroke
  fillColor?: string;
  height?: number;
  threshold?: number;
}

export const TimeSeriesChart: React.FC<TimeSeriesChartProps> = ({
  title,
  data,
  unit = '',
  color = '#3b82f6', // blue-500
  fillColor = 'rgba(59, 130, 246, 0.15)',
  height = 160,
  threshold,
}) => {
  if (!data || data.length === 0) {
    return (
      <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
        <h4 className="text-sm font-medium text-gray-300 mb-2">{title}</h4>
        <div className="h-32 flex items-center justify-center text-xs text-gray-500">
          No time-series data yet (gathering metrics...)
        </div>
      </div>
    );
  }

  const values = data.map((d) => d.value);
  const minVal = Math.min(...values);
  const maxVal = Math.max(...values, threshold ? threshold : minVal);
  const range = maxVal - minVal === 0 ? 1 : maxVal - minVal;

  const width = 450;
  const paddingX = 35;
  const paddingY = 20;
  const chartWidth = width - paddingX * 2;
  const chartHeight = height - paddingY * 2;

  const points = data.map((d, i) => {
    const x = paddingX + (i / Math.max(1, data.length - 1)) * chartWidth;
    const y = paddingY + chartHeight - ((d.value - minVal) / range) * chartHeight;
    return { x, y, ...d };
  });

  const pathD = points.reduce((acc, p, i) => {
    return i === 0 ? `M ${p.x} ${p.y}` : `${acc} L ${p.x} ${p.y}`;
  }, '');

  const areaD = `${pathD} L ${points[points.length - 1].x} ${paddingY + chartHeight} L ${points[0].x} ${paddingY + chartHeight} Z`;

  const latest = data[data.length - 1].value;
  const avg = values.reduce((a, b) => a + b, 0) / values.length;

  return (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
      <div className="flex justify-between items-start mb-2">
        <div>
          <h4 className="text-sm font-medium text-gray-300">{title}</h4>
          <span className="text-xl font-bold text-white">
            {latest.toFixed(1)} {unit}
          </span>
        </div>
        <div className="text-right text-xs text-gray-400">
          <span>Avg: {avg.toFixed(1)} {unit}</span>
          <span className="ml-3">Max: {maxVal.toFixed(1)} {unit}</span>
        </div>
      </div>

      <div className="relative">
        <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-auto overflow-visible">
          <defs>
            <linearGradient id={`grad-${title.replace(/\s+/g, '')}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={color} stopOpacity="0.3" />
              <stop offset="100%" stopColor={color} stopOpacity="0.0" />
            </linearGradient>
          </defs>

          <line
            x1={paddingX}
            y1={paddingY}
            x2={width - paddingX}
            y2={paddingY}
            stroke="#374151"
            strokeDasharray="2,2"
          />
          <line
            x1={paddingX}
            y1={paddingY + chartHeight / 2}
            x2={width - paddingX}
            y2={paddingY + chartHeight / 2}
            stroke="#374151"
            strokeDasharray="2,2"
          />
          <line
            x1={paddingX}
            y1={paddingY + chartHeight}
            x2={width - paddingX}
            y2={paddingY + chartHeight}
            stroke="#4b5563"
          />

          <text x={paddingX - 5} y={paddingY + 4} fill="#9ca3af" fontSize="9" textAnchor="end">
            {maxVal.toFixed(0)}
          </text>
          <text x={paddingX - 5} y={paddingY + chartHeight + 3} fill="#9ca3af" fontSize="9" textAnchor="end">
            {minVal.toFixed(0)}
          </text>

          {threshold && threshold <= maxVal && threshold >= minVal && (
            <line
              x1={paddingX}
              y1={paddingY + chartHeight - ((threshold - minVal) / range) * chartHeight}
              x2={width - paddingX}
              y2={paddingY + chartHeight - ((threshold - minVal) / range) * chartHeight}
              stroke="#ef4444"
              strokeDasharray="3,3"
            />
          )}

          <path d={areaD} fill={`url(#grad-${title.replace(/\s+/g, '')})`} />

          <path d={pathD} fill="none" stroke={color} strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" />

          {points.map((p, i) => (
            <circle
              key={i}
              cx={p.x}
              cy={p.y}
              r={i === points.length - 1 ? "4" : "2"}
              fill={i === points.length - 1 ? "#ffffff" : color}
              stroke={color}
              strokeWidth="1.5"
            />
          ))}
        </svg>

        <div className="flex justify-between text-[10px] text-gray-500 mt-1 px-1">
          <span>{data[0]?.time}</span>
          <span>{data[Math.floor(data.length / 2)]?.time}</span>
          <span>{data[data.length - 1]?.time}</span>
        </div>
      </div>
    </div>
  );
};
