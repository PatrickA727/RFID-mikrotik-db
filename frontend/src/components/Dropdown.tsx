import { FaChevronDown } from "react-icons/fa"; 
import { useNavigate } from "react-router-dom";
import { useState, useRef, useEffect } from "react";

// interface DropdownButtonProps {
//     isOpen: boolean;
//     toggleDropdown: () => void;
//   }

const Dropdown = () => {
    const [isOpen, setIsOpen] = useState(false)
    const dropdownRef = useRef<HTMLDivElement>(null)

    useEffect(() => {
        function handleClickOutside(event: MouseEvent) {
          if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
            setIsOpen(false)
          }
        }
    
        document.addEventListener('mousedown', handleClickOutside)
        return () => {
          document.removeEventListener('mousedown', handleClickOutside)
        }
    }, [])

    const toggleDropdown = () => {
        setIsOpen(!isOpen);
    }

    const navigate = useNavigate();
    const handleNavigation = (route: string) => {
        navigate(route);
        toggleDropdown();
    }

  return (
    <div className="relative inline-block text-left" ref={dropdownRef}>
      <button
        className="flex items-center justify-between w-full px-6 py-2 bg-white border border-gray-300 shadow-sm rounded-md hover:bg-gray-300 focus:outline-none"
        onClick={toggleDropdown}
      >
        Options
        <FaChevronDown className="ml-3" />
      </button>

      {isOpen && (
        <div className="absolute right-0 z-10 w-56 mt-1 origin-top-right bg-white border border-gray-300 divide-y divide-gray-100 rounded-md shadow-lg">
          <ul className="py-1">
            <li>
              <button
                onClick={() => handleNavigation("/type")}
                className="block w-full px-4 py-2 text-sm text-left text-gray-700 hover:bg-gray-100"
              >
                Types
              </button>

              <button
                onClick={() => handleNavigation("/sell")}
                className="block w-full px-4 py-2 text-sm text-left text-gray-700 hover:bg-gray-100"
              >
                Sell
              </button>
            </li>
          </ul>
        </div>
      )}
    </div>
  )
}

export default Dropdown