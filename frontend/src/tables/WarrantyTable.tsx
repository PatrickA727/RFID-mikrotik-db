import axios from 'axios';
import { useQuery } from "@tanstack/react-query";
import { createColumnHelper, flexRender, getCoreRowModel, getFilteredRowModel, getPaginationRowModel, useReactTable } from "@tanstack/react-table"
import { Button, ButtonGroup, Text } from "@chakra-ui/react";
import { useState } from "react";

interface Warranty {
    item_sn: number,
    purchase_date: Date,
    expiration: Date,
    cust_name: string,
    cust_email: string,
    cust_phone: string
}



const WarrantyTable = () => {
    const [globalFilter, setGlobalFilter] = useState("")

    const {data: warranties, error, isError, isLoading} = useQuery<Warranty[]>({
        queryKey: ["warranty"],
        queryFn: async(): Promise<Warranty[]> => {
            const { data } = await axios.get<Warranty[]>(`/api/item/get-warranties?limit=100&offset=0&search=`);
            return data
        }
    });

    const columnHelper = createColumnHelper<Warranty>()

    const columns = [
        columnHelper.accessor("item_sn", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Serial Number
                </span>
            )
        }),

        columnHelper.accessor("purchase_date", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Purchase Date
                </span>
            )
        }),

        columnHelper.accessor("expiration", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Expiration
                </span>
            )
        }),

        columnHelper.accessor("cust_name", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Customer Name
                </span>
            )
        }),

        columnHelper.accessor("cust_email", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Customer Email
                </span>
            )
        }),

        columnHelper.accessor("cust_phone", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Customer Phone
                </span>
            )
        }),
    ]

    const table = useReactTable({
        data: warranties || [],
        columns,
        state: {
            globalFilter
        },
        getPaginationRowModel: getPaginationRowModel(),
        getCoreRowModel: getCoreRowModel(),

        onGlobalFilterChange: setGlobalFilter,
        getFilteredRowModel: getFilteredRowModel(),
    })

    if (isLoading) {
        return <div>Loading...</div>;
      }
    
    if (isError) {
    return <div>Error: {(error as Error).message}</div>;
    }

    return (
        <div className="flex flex-col min-h-screen max-w-6xl mx-auto py-2 px-4 sm:px-6 lg:px-5">

            <div className="mb-4 relative">
                <input value={globalFilter ?? ""} onChange={(e) => setGlobalFilter(e.target.value)} placeholder="Search..." className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-gray-300 focus:border-gray-300 focus:outline-none"/>
            </div>

            <div className="overflow-x-auto bg-white shadow-md rounded-lg">
                <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                        {
                            table.getHeaderGroups().map((headerGroup) => (
                                <tr key={headerGroup.id}>
                                    {headerGroup.headers.map((header) => (
                                        <th key={header.id} className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            <div>
                                                {flexRender(
                                                    header.column.columnDef.header,
                                                    header.getContext()
                                                )}
                                            </div>
                                        </th>
                                    ))}
                                </tr>
                            ))
                        }
                    </thead>
                        <tbody className="bg-white divide-y divide-gray-200">
                            {table.getRowModel().rows.map((row) => (
                                <tr key={row.id} className="hover:bg-gray-50">
                                    {row.getVisibleCells().map((cell) => (
                                        <td key={cell.id} className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                        </td>
                                    ))}
                                </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <Text mb={2} className="mt-2">
                Page {table.getState().pagination.pageIndex + 1} of{" "}
                {table.getPageCount()}
            </Text>
            <ButtonGroup size="sm" isAttached variant="outline">
                <Button onClick={() => table.previousPage()} isDisabled={!table.getCanPreviousPage()} className="mr-1 p-2 rounded-md bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-50">
                    {"<"}
                </Button>
                <Button onClick={() => table.nextPage()} isDisabled={!table.getCanNextPage()} className="p-2 rounded-md bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-50">
                    {">"}
                </Button>
            </ButtonGroup>
        </div>
    )
}

export default WarrantyTable