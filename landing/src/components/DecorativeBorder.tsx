interface DecorativeBorderProps {
    side: 'left' | 'right';
}

export default function DecorativeBorder({ side }: DecorativeBorderProps) {
    const positionClass = side === 'left' ? 'left-6' : 'right-6';

    return (
        <div className={`fixed top-6 bottom-6 ${positionClass} hidden lg:flex flex-col items-center justify-between`}>
            <div className="top-0 bottom-0 absolute border-r border-current"></div>
            <div className="border-t w-3"></div>
            <div className="border-t w-3"></div>
            <div className="border-t w-3"></div>
            <div className="border-t w-3"></div>
            <div className="border-t w-3"></div>
            <div className="border-t w-3"></div>
            <div className="border-t w-3"></div>
        </div>
    );
}
