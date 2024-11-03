import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { Button, ButtonGroup, Text } from "@chakra-ui/react";
import { createColumnHelper, flexRender, getCoreRowModel, useReactTable } from "@tanstack/react-table"
import axios from 'axios';
import { useState, useCallback } from "react";
import { debounce } from 'lodash';

interface Item {
  serial_number: number,
  rfid_tag: string,
  item_name: string,
  warranty: string,
  sold: boolean,
  modal: number,
  keuntungan: number,
  quantity: number,
  batch: number,
  createdat: Date,
}

const TableScreen = () => {
    const [globalFilter, setGlobalFilter] = useState("")
    const [page, setPage] = useState(1)
    const [search, setSearch] = useState("") 

    const limit: number = 10
    let offset: number = 0

    if (page > 1 ) {
        offset = (page - 1) * limit 
    }

    const debouncedSetSearch = useCallback(
        debounce((query) => setSearch(query), 25),
        []
    );

    const handleSearchChange = (e) => {
        debouncedSetSearch(e.target.value);
    };
    
    const { data, error, isLoading, isError } = useQuery<{ items: Item[], item_count: number }>({
        queryKey: ['items', offset, search],    // Refetches when offset/search changes value
        queryFn: async (): Promise<{ items: Item[], item_count: number }> => { 
            const { data } = await axios.get<{ items: Item[], item_count: number }>(`/api/item/get-items?limit=${limit}&offset=${offset}&search=${search}`);
            console.log("API HIT: ", data)
            return data;
        },
        placeholderData: keepPreviousData,
    });

    // const updateItem = async (updatedItem: Item): Promise<Item | null> => {
    //     try {
    //         const response = await axios.patch<Item>(
    //             "/api/item/edit-item-sold", 
    //             updatedItem, 
    //             {
    //                 headers: {
    //                     'Content-Type': 'application/json',
    //                 }
    //             }
    //         );
    //         return response.data; // Return updated item
    //     } catch (error) {
    //         console.error("Error updating item:", error);
    //         return null; // Return null on failure
    //     }
    // }

    // const { mutate: updateItemMutation } = useMutation({
    //     mutationFn: updateItem,
    //     onSuccess: (data) => {
    //         console.log("Item updated successfully:", data);
    //     },
    //     onError: (error) => {
    //         console.error("Error updating item:", error.message);
    //     },
    // });

    const totalRecords: number = data?.item_count ?? 0;
    const canPrevPage: boolean = page > 1;
    const canNextPage: boolean = totalRecords > limit * page;
    let totalPages: number = Math.ceil(totalRecords / limit);
    if (totalPages < 1) {
        totalPages = 1
    }

    const columnHelper = createColumnHelper<Item>()

    const columns = [
        columnHelper.accessor("batch", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Batch
                </span>
            )
        }),

        columnHelper.accessor("createdat", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Date In
                </span>
            )
        }),

        columnHelper.accessor("serial_number", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Serial Number
                </span>
            )
        }),

        columnHelper.accessor("rfid_tag", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    RFID Tag
                </span>
            )
        }),

        columnHelper.accessor("item_name", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Item Name
                </span>
            )
        }),

        columnHelper.accessor("warranty", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Warranty
                </span>
            )
        }),

        columnHelper.accessor("sold", {
            cell: (info) => (info.getValue() ? 'Yes' : 'No'),
            header: () => (
                <span className="flex items-center">
                    Sold
                </span>
            )
        }),

        columnHelper.accessor("modal", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Modal
                </span>
            )
        }),

        columnHelper.accessor("keuntungan", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Keuntungan
                </span>
            )
        }),

        columnHelper.accessor("quantity", {
            cell: (info) => (
                info.getValue()
            ),
            header: () => (
                <span className="flex items-center">
                    Quantity
                </span>
            )
        }),
    ]

    const items = data?.items || [];

    const table = useReactTable({
        data: items,
        columns,
        state: {
            globalFilter
        },
        getCoreRowModel: getCoreRowModel(),
    })

    if (isLoading) {
        return <div>Loading...</div>;
      }
    
    if (isError) {
    return <div>Error: {(error as Error).message}</div>;
    }

    

    return (
        <div className="flex flex-col min-h-screen max-w-7xl mx-auto py-2 px-4 sm:px-6 lg:px-5">

            <div className="mb-4 relative">
                <input value={search ?? ""} onChange={handleSearchChange} placeholder="Search..." className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-gray-300 focus:border-gray-300 focus:outline-none"/>
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
                Page {page} of{" "}
                {totalPages}
            </Text>
            <ButtonGroup size="sm" isAttached variant="outline">
                <Button onClick={() => setPage(page - 1)} isDisabled={!canPrevPage} className="mr-1 p-2 rounded-md bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-50">
                    {"<"}
                </Button>
                <Button onClick={() => setPage(page + 1)} isDisabled={!canNextPage} className="p-2 rounded-md bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-50">
                    {">"}
                </Button>
            </ButtonGroup>
        </div>
    )
}

export default TableScreen