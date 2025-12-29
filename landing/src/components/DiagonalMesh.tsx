interface DiagonalMeshProps {
    className?: string;
}

export default function DiagonalMesh({ className = "" }: DiagonalMeshProps) {
    const patternId = `diagonal-mesh-${Math.random().toString(36).substr(2, 9)}`;

    return (
        <svg width="100%" height="100%" aria-hidden="true" className={className}>
            <defs>
                <pattern
                    viewBox="0 0 10 10"
                    width="10"
                    height="10"
                    patternUnits="userSpaceOnUse"
                    id={patternId}
                >
                    <line
                        x1="0"
                        y1="10"
                        x2="10"
                        y2="0"
                        stroke="currentColor"
                        vectorEffect="non-scaling-stroke"
                        className="stroke-black dark:stroke-white/30"
                    />
                </pattern>
            </defs>
            <rect width="100%" height="100%" fill={`url(#${patternId})`} />
        </svg>
    );
}
