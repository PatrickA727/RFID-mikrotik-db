// import { useState } from "react";
// import ItemSoldTable from "./tables/ItemSoldTable"
// import TableScreen from "./tables/TableScreen"
// import WarrantyTable from "./tables/WarrantyTable"
// import Dropdown from "./components/Dropdown";
import { Outlet } from 'react-router-dom';

function App() {
//   const [activeTab, setActiveTab] = useState('table1');

//   const renderTable = () => {
//     switch (activeTab) {
//         case 'table1':
//             return <TableScreen></TableScreen>;
//         case 'table2':
//             return <WarrantyTable></WarrantyTable>;
//         case 'table3':
//             return <ItemSoldTable></ItemSoldTable>;
//         default:
//             return <TableScreen></TableScreen>;
//     }
// };

  return (
    <>
    <Outlet></Outlet>
      {/* <div className="bg-gray-200">
        <div className="tab-bar flex justify-center space-x-4 py-3">
                  <Dropdown></Dropdown>
                  <button className="px-6 py-2 bg-white border border-gray-300 rounded-md shadow-sm " onClick={() => setActiveTab('table1')}>Items</button>
                  <button className="px-6 py-2 bg-white border border-gray-300 rounded-md shadow-sm " onClick={() => setActiveTab('table2')}>Warranty</button>
                  <button className="px-6 py-2 bg-white border border-gray-300 rounded-md shadow-sm " onClick={() => setActiveTab('table3')}>Sold Items</button>
        </div>

        <div className="table-content">
            {renderTable()}
        </div>
      </div> */}
    </>
  )
}

export default App
